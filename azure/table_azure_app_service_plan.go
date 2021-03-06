package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2020-06-01/web"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION ////

func tableAzureAppServicePlan(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azure_app_service_plan",
		Description: "Azure App Service Plan",
		Get: &plugin.GetConfig{
			KeyColumns:        plugin.AllColumns([]string{"name", "resource_group"}),
			Hydrate:           getAppServicePlan,
			ShouldIgnoreError: isNotFoundError([]string{"ResourceNotFound", "ResourceGroupNotFound"}),
		},
		List: &plugin.ListConfig{
			Hydrate: listAppServicePlans,
		},
		Columns: []*plugin.Column{
			{
				Name:        "name",
				Type:        proto.ColumnType_STRING,
				Description: "The friendly name that identifies the app service plan",
			},
			{
				Name:        "id",
				Description: "Contains ID to identify an app service plan uniquely",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "kind",
				Description: "Contains the kind of the resource",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "type",
				Description: "The resource type of the app service plan",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "hyper_v",
				Description: "Specify whether resource is Hyper-V container app service plan",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("AppServicePlanProperties.HyperV"),
				Default:     false,
			},
			{
				Name:        "is_spot",
				Description: "Specify whether this App Service Plan owns spot instances, or not",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("AppServicePlanProperties.IsSpot"),
				Default:     false,
			},
			{
				Name:        "is_xenon",
				Description: "Specify whether resource is Hyper-V container app service plan",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("AppServicePlanProperties.IsXenon"),
				Default:     false,
			},
			{
				Name:        "maximum_elastic_worker_count",
				Description: "Maximum number of total workers allowed for this ElasticScaleEnabled App Service Plan",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("AppServicePlanProperties.MaximumElasticWorkerCount"),
			},
			{
				Name:        "maximum_number_of_workers",
				Description: "Maximum number of instances that can be assigned to this App Service plan",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("AppServicePlanProperties.MaximumNumberOfWorkers"),
			},
			{
				Name:        "per_site_scaling",
				Description: "Specify whether apps assigned to this App Service plan can be scaled independently",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("AppServicePlanProperties.PerSiteScaling"),
				Default:     false,
			},
			{
				Name:        "provisioning_state",
				Description: "Provisioning state of the App Service Environment",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("AppServicePlanProperties.ProvisioningState").Transform(transform.ToString),
			},
			{
				Name:        "reserved",
				Description: "Specify whether the resource is Linux app service plan, or not",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("AppServicePlanProperties.Reserved"),
				Default:     false,
			},
			{
				Name:        "sku_capacity",
				Description: "Current number of instances assigned to the resource.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("Sku.Capacity"),
			},
			{
				Name:        "sku_family",
				Description: "Family code of the resource SKU",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Sku.Family"),
			},
			{
				Name:        "sku_name",
				Description: "Name of the resource SKU",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Sku.Name"),
			},
			{
				Name:        "sku_size",
				Description: "Size specifier of the resource SKU",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Sku.Size"),
			},
			{
				Name:        "sku_tier",
				Description: "Service tier of the resource SKU",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Sku.Tier"),
			},
			{
				Name:        "status",
				Description: "App Service plan status",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("AppServicePlanProperties.Status").Transform(transform.ToString),
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
				Description: "The name of the resource group in which the app service plan is created",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("AppServicePlanProperties.ResourceGroup"),
			},
			{
				Name:        "subscription_id",
				Description: "The Azure Subscription ID in which the resource is located",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("AppServicePlanProperties.Subscription"),
			},
		},
	}
}

//// FETCH FUNCTIONS ////

func listAppServicePlans(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	session, err := GetNewSession(ctx, d.ConnectionManager, "MANAGEMENT")
	if err != nil {
		return nil, err
	}
	subscriptionID := session.SubscriptionID

	webClient := web.NewAppServicePlansClient(subscriptionID)
	webClient.Authorizer = session.Authorizer

	pagesLeft := true
	for pagesLeft {
		param := true
		result, err := webClient.List(ctx, &param)
		if err != nil {
			return nil, err
		}

		for _, servicePlan := range result.Values() {
			d.StreamListItem(ctx, servicePlan)
		}
		result.NextWithContext(context.Background())
		pagesLeft = result.NotDone()
	}
	return nil, err
}

//// HYDRATE FUNCTIONS ////

func getAppServicePlan(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("getAppServicePlan")

	name := d.KeyColumnQuals["name"].GetStringValue()
	resourceGroup := d.KeyColumnQuals["resource_group"].GetStringValue()

	// resourceGroupName can't be empty
	// Error: pq: rpc error: code = Unknown desc = web.AppServicePlansClient#Get: Invalid input: autorest/validation: validation failed: parameter=resourceGroupName
	// constraint=MinLength value="" details: value length must be greater than or equal to 1
	if len(resourceGroup) < 1 {
		return nil, nil
	}

	session, err := GetNewSession(ctx, d.ConnectionManager, "MANAGEMENT")
	if err != nil {
		return nil, err
	}
	subscriptionID := session.SubscriptionID

	webClient := web.NewAppServicePlansClient(subscriptionID)
	webClient.Authorizer = session.Authorizer

	op, err := webClient.Get(ctx, resourceGroup, name)
	if err != nil {
		return nil, err
	}
	return op, nil
}
