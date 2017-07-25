BUILD_DIR=/go/src/github.com/twz123/kafro2go
BUILDER_IMAGE=kafro2go-builder

# binaries
DOCKER=docker
DEP=dep

kafro2go: .$(BUILDER_IMAGE) Gopkg.lock $(shell find pkg/ cmd/ -type f -name \*.go -print)
	$(DOCKER) run --rm -v "$(shell pwd -P):$(BUILD_DIR):ro" -w "$(BUILD_DIR)" $(BUILDER_IMAGE) \
	go build -o /dev/stdout -a --ldflags '-linkmode external -extldflags "-static"' -tags static_all cmd/kafro2go.go > kafro2go
	chmod +x kafro2go

Gopkg.lock: Gopkg.toml
	$(DEP) ensure
	find vendor/ -type f -name \*_test.* -delete
# this file doesn't conform to golangs convetions for tests and pulls in the testing framework
	rm -f vendor/github.com/confluentinc/confluent-kafka-go/kafka/testhelpers.go
# apply a patch to goavro to flatten unions
	( cd vendor/github.com/karrick/goavro && patch ) < unionless.patch

clean:
	rm -f kafro2go .$(BUILDER_IMAGE)

.$(BUILDER_IMAGE): Dockerfile.build
	$(DOCKER) build --pull=true -t $(BUILDER_IMAGE) - < Dockerfile.build
	touch .$(BUILDER_IMAGE)
