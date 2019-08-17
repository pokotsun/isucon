all: isuda isutar

deps:
	go get github.com/go-sql-driver/mysql
	go get github.com/gorilla/mux
	go get github.com/gorilla/sessions
	go get github.com/Songmu/strrand
	go get github.com/unrolled/render

isuda:
	go build -o isuda ./src/isuda/isuda.go ./src/isuda/type.go ./src/isuda/util.go ./src/isuda/entry.go ./src/isuda/star.go ./src/isuda/cache.go


isutar:
	go build -o isutar ./src/isuda/isutar.go ./src/isuda/type.go ./src/isuda/util.go ./src/isuda/cache.go

.PHONY: all deps
