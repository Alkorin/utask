NG=ng
NPM=npm
NODE=node
TAR=tar -zcvf
UNTAR=tar -zxvf
ECHO=echo
DEL=rm -rf

DIST_DIR=dist
DIST_TAR=dist.tar.gz
DIR_NODEMODULES=node_modules

BASEHREF ?= /

all: debug clean install build

clean:
	$(DEL) $(DIR_NODEMODULES)
	$(DEL) $(DIST_DIR)
	$(DEL) $(DIST_TAR)

debug: 
	$(NPM) -v
	$(NODE) -v
	$(ECHO) "-=DEBUG=-"

install:
	$(NPM) install

dev:
	$(NODE) scripts/prebuild-env-dev.js
	$(NG) serve --open --port 5557 --base-href $(BASEHREF)

build: install
	$(NODE) scripts/prebuild-env-dev.js
	$(NG) build --base-href $(BASEHREF)
	# $(TAR) $(DIST_TAR) $(DIST_DIR)

build-prod: install
	$(NODE) scripts/prebuild-env-prod.js
	$(NG) build --prod --base-href $(BASEHREF)
	# $(TAR) $(DIST_TAR) $(DIST_DIR)

prod:
	$(NODE) scripts/prebuild-env-prod.js
	$(NG) serve --open --port 5557 --base-href $(BASEHREF) --prod

package:
	$(UNTAR) $(DIST_TAR)

test:
	$(ECHO) "No Test for the moment"
