# HLF-2.2


# Build

```bash
# Original 
make native
make orderer-docker peer-docker tools-docker
# disable VSCC
make GO_TAGS="vscc" native
make orderer-docker peer-docker tools-docker

# disable vscc
GO_TAGS="vscc"
# sequential execute at endorser
GO_TAGS="redis"
# disable vscc & sequential execute
GO_TAGS="vscc redis"

# block-level parallel VSCC 
GO_TAGS="parallel"
# pipeline orderer?
GO_TAGS="pipeline"
```


