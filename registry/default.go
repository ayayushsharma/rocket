package registry

var DefaultRegistriesData string

func init() {
	DefaultRegistriesData = `
# These are the Rocket registries
#
# Lines starting with '#' are comments
#
# You can also use
# - Remote HTTP registries 
# - Local path registries
#
# Each registry must follow a schema
# Visit schema.URL

# Registries images are prioritised in the order the registries appear in this file
# If matching container image and versions are found, registry that appears
# first here only appear in the selection panel


# ayayushsharma github gist
https://gist.githubusercontent.com/ayayushsharma/da7d4e4bac0746879cca013f9225d391/raw/rocket.registry.json

# local registry
# /local/file/rocket/rocket.registry.json
`

}
