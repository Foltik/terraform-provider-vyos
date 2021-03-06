package vyos

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/foltik/vyos-client-go/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("VYOS_KEY", nil),
			},
			"cert": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"save": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				Description: "Save after making changes in Vyos",
			},
			"save_file": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "File to save configuration. Uses config.boot by default.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"vyos_config":              resourceConfig(),
			"vyos_config_block":        resourceConfigBlock(),
			"vyos_config_block_tree":   resourceConfigBlockTree(),
			"vyos_static_host_mapping": resourceStaticHostMapping(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"vyos_config": dataSourceConfig(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

type ProviderClass struct{
	schema *schema.ResourceData;
	client *client.Client;
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	url := d.Get("url").(string)
	key := d.Get("key").(string)

	cert := d.Get("cert").(string)
	c := &client.Client{}

	if cert != "" {
		return nil, diag.Errorf("TODO: Use trusted self signed certificate")
	} else {
		// Just allow self signed certificates if a trusted cert isn't specified
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		cc := &http.Client{Transport: tr, Timeout: 10 * time.Minute}
		c = client.NewWithClient(cc, url, key)
	}

	return &ProviderClass{d, c}, diag.Diagnostics{}
}

func (p *ProviderClass) conditionalSave(ctx context.Context) {
	save      := p.schema.Get("save").(bool)
	save_file := p.schema.Get("save_file").(string)

	if (save) {
		if save_file == "" {
			p.client.Config.Save(ctx);
		} else {
			p.client.Config.SaveFile(ctx, save_file);
		}
	}
}
