all: isuda

deps:
	go get github.com/go-sql-driver/mysql
	go get github.com/gorilla/mux
	go get github.com/gorilla/sessions
	go get github.com/Songmu/strrand
	go get github.com/unrolled/render
	go get github.com/patrickmn/go-cache


isuda: isuda.go type.go util.go htmlify.go cache.go
	go build -o isuda isuda.go type.go util.go htmlify.go cache.go

#isutar: deps
#	go build -o isutar isutar.go type.go util.go

.PHONY: all deps
