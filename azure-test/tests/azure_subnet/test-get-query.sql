select name, id, virtual_network_name, type, address_prefix, delegations, resource_group
from azure.azure_subnet
where name = '{{resourceName}}' and resource_group = '{{resourceName}}' and virtual_network_name = '{{resourceName}}'
