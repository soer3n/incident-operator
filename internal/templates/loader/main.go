package loader

import (
	"bufio"
	"bytes"
	b64 "encoding/base64"
	"errors"
	"io"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"
	kyaml "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/sirupsen/logrus"
	"github.com/soer3n/incident-operator/internal/templates/resources"
	"github.com/soer3n/incident-operator/internal/webhook"
)

const DEFAULT_NAMESPACE = "incident-operator"
const DEFAULT_SUBJECT = "quarantine-webhook." + DEFAULT_NAMESPACE + ".svc"

func LoadManifests(data interface{}, logger logrus.FieldLogger) ([]runtime.RawExtension, error) {

	fsys := resources.FS
	list := []runtime.RawExtension{}

	files, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}

	for _, file := range files {

		ext := strings.ToLower(filepath.Ext(file.Name()))
		// Only YAML, YML and JSON manifests are supported
		switch ext {
		case ".tmpl":
			logger.Infof("File: %v", file.Name())
		default:
			logger.Infof("Skipping file %q because it's not a template file\n", file.Name())

			continue
		}

		logger.Infof("Parsing manifests '%s'\n", file.Name())

		filePath := filepath.Join(".", file.Name())

		if file.IsDir() {
			continue
		}

		logger.Infoln("Parsing manifests to bytes")
		manifestBytes, err := fs.ReadFile(fsys, filePath)

		if err != nil {
			return nil, err
		}

		logger.Infoln("Init new template struct")
		tpl, err := template.New("deployment").Parse(string(manifestBytes))

		if err != nil {
			return nil, err
		}

		logger.Infoln("Rendering template variables")
		buf := bytes.NewBuffer([]byte{})

		if err := tpl.Execute(buf, data); err != nil {
			return nil, err
		}

		reader := kyaml.NewYAMLReader(bufio.NewReader(buf))

		for {
			b, err := reader.Read()

			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				return nil, err
			}

			b = bytes.TrimSpace(b)
			if len(b) == 0 {
				continue
			}

			decoder := kyaml.NewYAMLToJSONDecoder(bytes.NewBuffer(b))
			raw := runtime.RawExtension{}

			if err := decoder.Decode(&raw); err != nil {
				return nil, err
			}

			list = append(list, raw)
		}

	}

	return list, nil
}

type Config struct {
	Certs     Certs
	Namespace string
}

type Certs struct {
	Ca   string
	Cert string
	Key  string
}

func LoadConfig(path string, logger logrus.FieldLogger) (*Config, error) {

	c := &Config{}

	logger.Infoln("read config file")

	f, err := ioutil.ReadFile(path)

	if err != nil {

		logger.Infoln("could not found config file. Setting default values.")
		c.Namespace = DEFAULT_NAMESPACE

	}

	if f != nil {
		logger.Infoln("unmarshal bytes to config struct")

		if err := yaml.Unmarshal(f, c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Config) SetCerts(path string, logger logrus.FieldLogger) error {

	logger.Infoln("generating ca bundle...")
	wc := &webhook.Cert{}

	if err := wc.GenerateWebhookCert(); err != nil {
		return err
	}

	logger.Infoln("generating webhook server cert...")

	if err := wc.Create(DEFAULT_SUBJECT); err != nil {
		return err
	}

	c.Certs = Certs{
		Ca:   b64.URLEncoding.EncodeToString(wc.Ca.Cert),
		Cert: b64.URLEncoding.EncodeToString(wc.Cert),
		Key:  b64.URLEncoding.EncodeToString(wc.Key),
	}

	return nil
}
