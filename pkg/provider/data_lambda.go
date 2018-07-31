package provider

import (
	"archive/zip"
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/chanzuckerberg/terraform-provider-bless/pkg/util"
	"github.com/gobuffalo/packr"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
)

const (
	schemaServiceName  = "service_name"
	schemaKMSAuthKeyID = "kmsauth_key_id"

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
	// EncryptedPassword is the kms encrypted password for the CA private key
	EncryptedPassword string
	// EncryptedPrivateKey is a password encrypted CA private key
	EncryptedPrivateKey string
	// Name is the name of this service
	Name string
	// KMSAuthKeyID is the kmsauth key ID
	KMSAuthKeyID string
}

// resourceLambda is a namespace
type resourceLambda struct{}

func newResourceLambda() *resourceLambda {
	return &resourceLambda{}
}

func (l *resourceLambda) writeFileToZip(f io.Reader, fileInfo os.FileInfo, writer *zip.Writer, path string,
) error {
	relativeName, err := filepath.Rel("", path)
	if err != nil {
		return errors.Wrapf(err, "Could not create relative path %s for zip", path)
	}
	fh, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return errors.Wrapf(err, "Could not create zip file header for %s", relativeName)
	}
	fh.Name = filepath.ToSlash(relativeName)
	fh.Method = zip.Deflate
	fh.SetModTime(time.Time{})
	w, err := writer.CreateHeader(fh)
	if err != nil {
		return errors.Wrapf(err, "Could not create zip writer for %s", relativeName)
	}
	_, err = io.Copy(w, f)
	return errors.Wrapf(err, "Could not add file %s to zip", relativeName)
}

// getBlessConfig reads and templetizes a bless config
func (l *resourceLambda) getBlessConfig(d *schema.ResourceData) (io.Reader, os.FileInfo, error) {
	templateBox := packr.NewBox("../../bless_lambda")
	tpl, err := templateBox.Open("bless_deploy.cfg.tpl")
	if err != nil {
		return nil, nil, errors.Wrap(err, "Could not open pckr box for bless_deploy.cfg.tpl")
	}
	tplBytes, err := ioutil.ReadAll(tpl)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Could not read bless_deploy.cfg.tpl")
	}
	fileInfo, err := tpl.Stat()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Could not stat bless_deploy.cfg.tpl")
	}
	t, err := template.New("config").Parse(string(tplBytes))
	if err != nil {
		return nil, nil, errors.Wrap(err, "Could not load template")
	}
	blessConfig := blessConfig{
		EncryptedPassword:   d.Get(schemaEncryptedPassword).(string),
		EncryptedPrivateKey: d.Get(schemaEncryptedPrivateKey).(string),
		Name:                d.Get(schemaServiceName).(string),
		KMSAuthKeyID:        d.Get(schemaKMSAuthKeyID).(string),
	}
	buff := bytes.NewBuffer(nil)
	err = t.Execute(buff, blessConfig)
	return buff, fileInfo, errors.Wrap(err, "Could not templetize config")
}

// Create bundles the lambda code and configuration into a zip that can be uploaded to AWS lambda
func (l *resourceLambda) Read(d *schema.ResourceData, meta interface{}) error {
	path := d.Get(schemaOutputPath).(string)
	outFile, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "Could not open output file at %s", path)
	}
	defer outFile.Close()
	writer := zip.NewWriter(outFile)
	defer writer.Close()

	// Add all the python lambda files to the zip
	zipBox := packr.NewBox("../../bless_lambda/bless_ca")
	err = zipBox.Walk(func(path string, f packr.File) error {
		fileInfo, err := f.FileInfo()
		if err != nil {
			return errors.Wrapf(err, "Could not get file info for %s", path)
		}
		return l.writeFileToZip(f, fileInfo, writer, path)
	})

	blessConfig, blessConfigFileInfo, err := l.getBlessConfig(d)
	if err != nil {
		return err
	}
	// Write the config
	err = l.writeFileToZip(blessConfig, blessConfigFileInfo, writer, "bless_deploy.cfg")
	if err != nil {
		return err
	}

	// Calculate file hash for tf state
	fileHash, err := util.HashFileForState(outFile.Name())
	if err != nil {
		return err
	}

	d.Set(SchemaOutputBase64Sha256, fileHash)
	d.SetId(fileHash)
	return err
}
