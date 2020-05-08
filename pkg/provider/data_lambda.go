package provider

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/chanzuckerberg/terraform-provider-bless/pkg/util"
	"github.com/gobuffalo/packr"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
)

const (
	schemaServiceName                                   = "service_name"
	schemaLoggingLevel                                  = "logging_level"
	schemaUsernameValidation                            = "username_validation"
	schemaKMSAuthKeyID                                  = "kmsauth_key_id"
	schemaKMSAuthRemoteUsernamesAllowed                 = "kmsauth_remote_usernames_allowed"
	schemaKMSAuthValidateRemoteUsernameAgainstIAMGroups = "kmsauth_validate_remote_user"
	schemaKMSAuthIAMGroupNameFormat                     = "kmsauth_iam_group_name_format"

	// SchemaOutputBase64Sha256 is the base64 encoded sha256 of bless.zip contents
	SchemaOutputBase64Sha256 = "output_base64sha256"
	// schemaOutputPath is the output_path of the zip
	schemaOutputPath = "output_path"
)

// Lambda is a bless lambda resource
func Lambda() *schema.Resource {
	lambda := newResourceLambda()
	return &schema.Resource{
		Read: lambda.Read,

		Schema: map[string]*schema.Schema{
			schemaEncryptedPassword: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The kms encrypted password for the CA",
				ForceNew:    true,
			},
			schemaEncryptedPrivateKey: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The encrypted CA private key",
				ForceNew:    true,
			},
			schemaServiceName: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the bless CA service. Used for kmsauth.",
				ForceNew:    true,
			},
			schemaKMSAuthKeyID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The kmsauth key ID",
				ForceNew:    true,
			},
			schemaOutputPath: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Path where the bless zip archive will be written",
				ForceNew:    true,
			},
			schemaLoggingLevel: &schema.Schema{
				Type:        schema.TypeString,
				Default:     "INFO",
				Optional:    true,
				ForceNew:    true,
				Description: "Bless lambda logging level",
			},
			schemaUsernameValidation: &schema.Schema{
				Type:        schema.TypeString,
				Default:     "email",
				Optional:    true,
				ForceNew:    true,
				Description: "Bless lambda default username validation",
			},
			schemaKMSAuthRemoteUsernamesAllowed: &schema.Schema{
				Type:        schema.TypeString,
				Default:     "*",
				Optional:    true,
				ForceNew:    true,
				Description: "The remote usernames allowed. \"*\" indicates any",
			},
			schemaKMSAuthValidateRemoteUsernameAgainstIAMGroups: &schema.Schema{
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				ForceNew:    true,
				Description: "If bless should validate a remote username against an IAM group membership",
			},
			schemaKMSAuthIAMGroupNameFormat: &schema.Schema{
				Type:        schema.TypeString,
				Default:     "ssh-{}",
				Optional:    true,
				ForceNew:    true,
				Description: "The format of IAM Group Name used to validate membership.",
			},

			// computed
			SchemaOutputBase64Sha256: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Base64Sha256 or temporary bless.zip contents",
			},
		},
	}

}

//
type blessConfig struct {
	// Name is the name of this service
	Name string
	// LoggingLevel
	LoggingLevel string
	// UsernameValidation tells bless how to validate usernames
	UsernameValidation string
	// EncryptedPassword is the kms encrypted password for the CA private key
	EncryptedPassword string
	// EncryptedPrivateKey is a password encrypted CA private key
	EncryptedPrivateKey string
	// KMSAuthKeyID the kmsauth kms key id
	KMSAuthKeyID string
	// KMSAuthRemoteUsernamesAllowed the remote usernames allowed
	KMSAuthRemoteUsernamesAllowed string
	// KMSAuthValidateRemoteUsernameAgainstIAMGroups if kmsauth should validate the remote username against an IAM group membership
	KMSAuthValidateRemoteUsernameAgainstIAMGroups bool
	// KMSAuthIAMGroupNameFormat a pattern to fetch iam groups typically ssh-{} where {} will be replaced with the remote-username
	KMSAuthIAMGroupNameFormat string
}

// resourceLambda is a namespace
type resourceLambda struct{}

func newResourceLambda() *resourceLambda {
	return &resourceLambda{}
}

func (l *resourceLambda) writeFileToZip(f io.Reader, writer *zip.Writer, path string,
) error {
	relativeName, err := filepath.Rel("", path)
	if err != nil {
		return errors.Wrapf(err, "Could not create relative path %s for zip", path)
	}
	fh := &zip.FileHeader{}
	fh.Name = filepath.ToSlash(relativeName)
	fh.Method = zip.Deflate
	w, err := writer.CreateHeader(fh)
	if err != nil {
		return errors.Wrapf(err, "Could not create zip writer for %s", relativeName)
	}
	_, err = io.Copy(w, f)
	return errors.Wrapf(err, "Could not add file %s to zip", relativeName)
}

// getBlessConfig reads and templetizes a bless config
func (l *resourceLambda) getBlessConfig(d *schema.ResourceData) (io.Reader, error) {
	templateBox := packr.NewBox("../../bless_lambda")
	tpl, err := templateBox.Open("bless_deploy.cfg.tpl")
	if err != nil {
		return nil, errors.Wrap(err, "Could not open pckr box for bless_deploy.cfg.tpl")
	}
	tplBytes, err := ioutil.ReadAll(tpl)
	if err != nil {
		return nil, errors.Wrap(err, "Could not read bless_deploy.cfg.tpl")
	}
	t, err := template.
		New("config").
		Funcs(map[string]interface{}{
			"pythonBool": func(isTrue bool) string {
				if isTrue {
					return "True"
				}
				return "False"
			},
		}).
		Parse(string(tplBytes))

	if err != nil {
		return nil, errors.Wrap(err, "Could not load template")
	}
	blessConfig := blessConfig{
		Name:                          d.Get(schemaServiceName).(string),
		LoggingLevel:                  d.Get(schemaLoggingLevel).(string),
		UsernameValidation:            d.Get(schemaUsernameValidation).(string),
		EncryptedPassword:             d.Get(schemaEncryptedPassword).(string),
		EncryptedPrivateKey:           d.Get(schemaEncryptedPrivateKey).(string),
		KMSAuthKeyID:                  d.Get(schemaKMSAuthKeyID).(string),
		KMSAuthRemoteUsernamesAllowed: d.Get(schemaKMSAuthRemoteUsernamesAllowed).(string),
		KMSAuthValidateRemoteUsernameAgainstIAMGroups: d.Get(schemaKMSAuthValidateRemoteUsernameAgainstIAMGroups).(bool),
		KMSAuthIAMGroupNameFormat:                     d.Get(schemaKMSAuthIAMGroupNameFormat).(string),
	}

	buff := bytes.NewBuffer(nil)
	err = t.Execute(buff, blessConfig)
	return buff, errors.Wrap(err, "Could not templetize config")
}

// archive generates the zip archive
func (l *resourceLambda) archive(d *schema.ResourceData, meta interface{}) error {
	outputPath := d.Get(schemaOutputPath).(string)
	outputDirectory := path.Dir(outputPath)
	if outputDirectory != "" {
		if _, err := os.Stat(outputDirectory); err != nil {
			if err := os.MkdirAll(outputDirectory, 0755); err != nil {
				return errors.Wrapf(err, "Could not create directories %s", outputDirectory)
			}
		}
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return errors.Wrapf(err, "Could not open output file at %s", outputPath)
	}
	defer outFile.Close()
	writer := zip.NewWriter(outFile)
	defer writer.Close()

	// Add all the python lambda files to the zip
	zipBox := packr.NewBox("../../bless_lambda/bless_ca")
	// HACK: zipBox.Walk does not guarantee a stable iteration order
	files := []string{}
	err = zipBox.Walk(func(path string, f packr.File) error {
		fileInfo, err := f.FileInfo()
		if err != nil {
			return errors.Wrapf(err, "Could not get file info for %s", path)
		}
		if fileInfo.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "could not walk zip")
	}

	// Sort so stable adding of files to zip
	sort.Strings(files)
	for _, path := range files {
		f, err := zipBox.Open(path)
		if err != nil {
			return errors.Wrapf(err, "Could not open file %s", path)
		}
		err = l.writeFileToZip(f, writer, path)
		if err != nil {
			return err
		}
	}

	blessConfig, err := l.getBlessConfig(d)
	if err != nil {
		return err
	}

	// Write the config
	return l.writeFileToZip(blessConfig, writer, "bless_deploy.cfg")
}

// Create bundles the lambda code and configuration into a zip that can be uploaded to AWS lambda
func (l *resourceLambda) Read(d *schema.ResourceData, meta interface{}) error {
	outputPath := d.Get(schemaOutputPath).(string)
	err := l.archive(d, meta)
	if err != nil {
		return err
	}
	// Calculate file hash for tf state
	fileHash, err := util.HashFileForState(outputPath)
	if err != nil {
		return err
	}
	d.Set(SchemaOutputBase64Sha256, fileHash) //nolint
	d.SetId(fileHash) //nolint
	return err
}
