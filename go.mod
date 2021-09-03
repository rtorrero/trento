module github.com/trento-project/trento

go 1.16

require (
	github.com/aquasecurity/bench-common v0.4.4
	github.com/boj/redistore v0.0.0-20180917114910-cd5dcc76aeff // indirect
	github.com/cloudquery/sqlite v1.0.1
	github.com/gin-gonic/contrib v0.0.0-20201101042839-6a891bf89f19
	github.com/gin-gonic/gin v1.7.0
	github.com/go-playground/assert/v2 v2.0.1
	github.com/gomarkdown/markdown v0.0.0-20210514010506-3b9f47219fe7
	github.com/gorilla/sessions v1.2.1 // indirect
	github.com/hashicorp/consul-template v0.25.2
	github.com/hashicorp/consul/api v1.4.0
	github.com/hooklift/gowsdl v0.5.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/afero v1.1.2
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/tdewolff/minify/v2 v2.9.16
	github.com/vektra/mockery/v2 v2.9.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	gorm.io/driver/postgres v1.1.0
	gorm.io/gorm v1.21.14
)

replace github.com/trento-project/trento => ./
