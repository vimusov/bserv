target := "bserv"

default: build

build:
    go build -o {{target}} {{target}}.go

clean:
    rm -f {{target}}

install destdir: build
    install -D -m 0755 {{target}} "{{destdir}}"/usr/bin/{{target}}
