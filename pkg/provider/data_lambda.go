package provider

import (
	"archive/zip"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/chanzuckerberg/terraform-provider-bless/pkg/util"
	"github.com/gobuffalo/packr"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
)

const (
	// SchemaOutputPath is the output_path of the zip
	SchemaOutputPath         = "output_path"
	schemaServiceName        = "service_name"
	schemaOutputBase64Sha256 = "output_base64sha256"
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

			// computed
			SchemaOutputPath: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Temporary directory that holds the bless zip",
			},
			schemaOutputBase64Sha256: &schema.Schema{
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

// Create bundles the lambda code and configuration into a zip that can be uploaded to AWS lambda
func (l *resourceLambda) Read(d *schema.ResourceData, meta interface{}) error {
	outFile, err := ioutil.TempFile("", "bless.zip")
	if err != nil {
		return errors.Wrap(err, "Could not open temporary file")
	}
	defer outFile.Close()
	writer := zip.NewWriter(outFile)
	defer writer.Close()

	zipBox := packr.NewBox("../../bless_lambda/bless_ca")
	err = zipBox.Walk(func(path string, f packr.File) error {
		relname, err := filepath.Rel("", path)
		if err != nil {
			return err
		}
		fileInfo, err := f.FileInfo()
		if err != nil {
			return err
		}
		fh, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return err
		}
		fh.Name = filepath.ToSlash(relname)
		fh.Method = zip.Deflate
		// fh.Modified alone isn't enough when using a zero value
		fh.SetModTime(time.Time{})

		w, err := writer.CreateHeader(fh)
		if err != nil {
			return err
		}

		contents, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		_, err = w.Write(contents)
		if err != nil {
			return err
		}
		return nil
	})

	// Templ config
	templateBox := packr.NewBox("../../bless_lambda/bless_deploy.cfg.tpl")
	tpl, err := templateBox.Open("")
	if err != nil {
		return err
	}
	tplBytes, err := ioutil.ReadAll(tpl)
	if err != nil {
		return err
	}
	fileInfo, err := tpl.Stat()
	if err != nil {
		return err
	}

	fh, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}
	relname, err := filepath.Rel("", "bless_deploy.cfg")
	if err != nil {
		return err
	}

	fh.Name = filepath.ToSlash(relname)
	fh.Method = zip.Deflate

	t, err := template.New("config").Parse(string(tplBytes))
	if err != nil {
		return errors.Wrap(err, "could not load template")
	}

	w, err := writer.CreateHeader(fh)
	if err != nil {
		return err
	}

	blessConfig := blessConfig{
		EncryptedPassword:   d.Get(schemaEncryptedPassword).(string),
		EncryptedPrivateKey: d.Get(schemaEncryptedPrivateKey).(string),
		Name:                d.Get(schemaServiceName).(string),
		KMSAuthKeyID:        d.Get(schemaKmsKeyID).(string),
	}
	err = t.Execute(w, blessConfig)
	if err != nil {
		return err
	}

	fileHash, err := util.HashFileForState(outFile.Name())
	if err != nil {
		return err
	}

	d.Set(SchemaOutputPath, outFile.Name())
	d.Set(schemaOutputBase64Sha256, fileHash)
	d.SetId(fileHash)
	return err
}
