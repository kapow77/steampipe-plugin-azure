package azure

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2020-06-01/web"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION ////

func tableAzureAppServiceFunctionApp(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azure_app_service_function_app",
		Description: "Azure App Service Function App",
		Get: &plugin.GetConfig{
			KeyColumns:        plugin.AllColumns([]string{"name", "resource_group"}),
			Hydrate:           getAppServiceFunctionApp,
			ShouldIgnoreError: isNotFoundError([]string{"ResourceNotFound", "ResourceGroupNotFound"}),
		},
		List: &plugin.ListConfig{
			Hydrate: listAppServiceFunctionApps,
		},
		Columns: []*plugin.Column{
			{
				Name:        "name",
				Description: "The friendly name that identifies the app service function app",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "id",
				Description: "Contains ID to identify an app service function app uniquely",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "kind",
				Description: "Contains the kind of the resource",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "state",
				Description: "Current state of the app",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("SiteProperties.State"),
			},
			{
				Name:        "type",
				Description: "The resource type of the app service function app",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "client_affinity_enabled",
				Description: "Specify whether client affinity is enabled",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("SiteProperties.ClientAffinityEnabled"),
			},
			{
				Name:        "client_cert_enabled",
				Description: "Specify whether client certificate authentication is enabled",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("SiteProperties.ClientCertEnabled"),
			},
			{
				Name:        "default_site_hostname",
				Description: "Default hostname of the app",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("SiteProperties.DefaultHostName"),
			},
			{
				Name:        "enabled",
				Description: "Specify whether the app is enabled",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("SiteProperties.Enabled"),
			},
			{
				Name:        "host_name_disabled",
				Description: "Specify whether the public hostnames of the app is disabled",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("SiteProperties.HostNamesDisabled"),
			},
			{
				Name:        "https_only",
				Description: "Specify whether configuring a web site to accept only https requests",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("SiteProperties.HTTPSOnly"),
			},
			{
				Name:        "outbound_ip_addresses",
				Description: "List of IP addresses that the app uses for outbound connections (e.g. database access)",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("SiteProperties.OutboundIPAddresses"),
			},
			{
				Name:        "possible_outbound_ip_addresses",
				Description: "List of possible IP addresses that the app uses for outbound connections (e.g. database access)",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("SiteProperties.PossibleOutboundIPAddresses"),
			},
			{
				Name:        "reserved",
				Description: "Specify whether the app is reserved",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("SiteProperties.Reserved"),
			},
			{
				Name:        "host_names",
				Description: "A list of hostnames associated with the app",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("SiteProperties.HostNames"),
			},
			{
				Name:        "site_config",
				Description: "A map of all configuration for the app",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("SiteProperties.SiteConfig"),
			},

			// Standard columns
			{
				Name:        "title",
				Description: resourceInterfaceDescription("title"),
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "tags",
				Description: resourceInterfaceDescription("tags"),
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "akas",
				Description: resourceInterfaceDescription("akas"),
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("ID").Transform(idToAkas),
			},
			{
				Name:        "region",
				Description: "The Azure region in which the resource is located",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Location"),
			},
			{
				Name:        "resource_group",
				Description: "The name of the resource group in which the app service function app is created",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("SiteProperties.ResourceGroup"),
			},
			{
				Name:        "subscription_id",
				Description: "The Azure Subscription ID in which the resource is located",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID").Transform(idToSubscriptionID),
			},
		},
	}
}

//// FETCH FUNCTIONS ////

func listAppServiceFunctionApps(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	session, err := GetNewSession(ctx, d.ConnectionManager, "MANAGEMENT")
	if err != nil {
		return nil, err
	}
	subscriptionID := session.SubscriptionID

	webClient := web.NewAppsClient(subscriptionID)
	webClient.Authorizer = session.Authorizer

	pagesLeft := true
	for pagesLeft {
		result, err := webClient.List(ctx)
		if err != nil {
			return nil, err
		}

		for _, functionApp := range result.Values() {
			// Filtering out all the web apps
			if strings.Contains(string(*functionApp.Kind), "functionapp") {
				d.StreamListItem(ctx, functionApp)
			}
		}
		result.NextWithContext(context.Background())
		pagesLeft = result.NotDone()
	}
	return nil, err
}

//// HYDRATE FUNCTIONS ////

func getAppServiceFunctionApp(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("getAppServiceFunctionApp")

	name := d.KeyColumnQuals["name"].GetStringValue()
	resourceGroup := d.KeyColumnQuals["resource_group"].GetStringValue()

	// Error: pq: rpc error: code = Unknown desc = web.AppsClient#Get: Invalid input: autorest/validation: validation failed: parameter=resourceGroupName
	// constraint=MinLength value="" details: value length must be greater than or equal to 1
	if len(resourceGroup) < 1 {
		return nil, nil
	}

	session, err := GetNewSession(ctx, d.ConnectionManager, "MANAGEMENT")
	if err != nil {
		return nil, err
	}
	subscriptionID := session.SubscriptionID

	webClient := web.NewAppsClient(subscriptionID)
	webClient.Authorizer = session.Authorizer

	op, err := webClient.Get(ctx, resourceGroup, name)
	if err != nil {
		return nil, err
	}

	// In some cases resource does not give any notFound error
	// instead of notFound error, it returns empty data
	if op.ID != nil {
		return op, nil
	}

	return nil, nil
}
