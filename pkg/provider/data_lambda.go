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
	SchemaOutputPath = "output_path"
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

			// computed
			SchemaOutputPath: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Temporary directory that holds the bless zip",
			},
		},
	}

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

	relname, err := filepath.Rel("", templateBox.Path)
	if err != nil {
		return err
	}

	t, err := template.New("config").Parse(string(tplBytes))
	if err != nil {
		return errors.Wrap(err, "could not load template")
	}

	d.Set(SchemaOutputPath, outFile.Name())
	d.SetId(util.HashForState(outFile.Name()))
	return err
}
