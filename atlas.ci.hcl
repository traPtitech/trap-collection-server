lint {
  non_linear {
    error = true
  }
}

env "local" {
  dev = "docker://mariadb/10.7/trap_collection"
  
  migration {
    dir = "file://migrations"
  }

  lint {
    git {
      base = "origin/main"
    }
  }
}

env "ci" {
  dev = "mysql://root:pass@localhost:3306/trap_collection"

  lint {
    git {
      base = "origin/main"
    }
  }
}
