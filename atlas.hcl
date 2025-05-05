data "external_schema" "gorm" {
  program = [
    "go",
    "tool",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./src/repository/gorm2/schema",
    "--dialect", "mysql", // | postgres | sqlite | sqlserver
  ]
}

lint {
  non_linear {
    error = true
  }
}

env "local" {
  src = data.external_schema.gorm.url
  dev = "docker://mariadb/10.7/trap_collection"
  url = "mysql://root:pass@localhost:3306/trap_collection"
  
  migration {
    dir = "file://migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

env "ci" {
  dev = "mysql://root:pass@localhost:3306/trap_collection"

  lint {
    git {
      base = "main"
    }
  }
}
