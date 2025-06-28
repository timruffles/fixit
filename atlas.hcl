env "local" {
  src = "ent://engine/ent/schema"
  dev = "docker://postgres/15/fixit?search_path=public"
  migration {
    dir = "file://engine/ent/migrate/migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

env "docker" {
  src = "ent://engine/ent/schema"
  url = "postgres://fixit:password@localhost:5432/fixit?sslmode=disable"
  migration {
    dir = "file://engine/ent/migrate/migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}