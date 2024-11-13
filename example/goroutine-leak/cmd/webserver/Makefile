# --go_opt=Mproto/notification/v3/models=./gen/go/proto/notification/v3/models <- go_package
# find proto -type f -name '*.proto' -exec protoc -I=$PWD -I$PWD/proto --go_out=$PWD/gen/go --go_opt=paths=source_relative --go_opt=Mproto/notification/v3/models=./gen/go/proto/notification/v3/models {} \;
TWIRP_OUT=gen/twirp
CONNECT_OUT=gen/connect

GEN_OUT=./gen

clean:
	rm -rf gen
	mkdir -p ${GEN_OUT}
.PHONY: clean

twirp:
	protoc \
	-I ./proto \
	--go_out=${GEN_OUT} \
	--twirp_out=${GEN_OUT} \
	proto/user/v1/user.proto
.PHONY: twirp

connect:
	protoc \
	-I ./proto \
	--go_out=${GEN_OUT} \
	--connect-go_out=${GEN_OUT} \
	proto/user/v1/user.proto
.PHONY: connect

buf_update:
	buf dep update

buf_gen:
	buf generate