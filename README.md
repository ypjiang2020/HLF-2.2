# HLF-2.2


# Build

```bash
# Original 
make native
make orderer-docker peer-docker tools-docker
# disable VSCC
make GO_TAGS="vscc" native
make orderer-docker peer-docker tools-docker
```



