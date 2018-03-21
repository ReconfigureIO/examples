# variable definitions
.PHONY: test test-real 

all: test 

test: 
	@tmpdir=`mktemp -d`; \
	trap 'rm -rf "$$tmpdir"' EXIT; \
	$(MAKE) test-real TMPDIR=$$tmpdir 

test-real:
	curl -Lo $(TMPDIR)/reco.zip https://s3.amazonaws.com/reconfigure.io/reco/releases/reco-v0.5.1-x86_64-linux.zip 
	unzip $(TMPDIR)/reco.zip -d $(TMPDIR)
	$(TMPDIR)/reco version
	$(TMPDIR)/reco check --source $(shell pwd)/addition
	$(TMPDIR)/reco check --source $(shell pwd)/histogram-array
	$(TMPDIR)/reco check --source $(shell pwd)/histogram-array-SMI
	$(TMPDIR)/reco check --source $(shell pwd)/histogram-parallel
	$(TMPDIR)/reco check --source $(shell pwd)/memcopy
	$(TMPDIR)/reco check --source $(shell pwd)/memtest	
