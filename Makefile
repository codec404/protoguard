.PHONY: testdata-proto
testdata-proto:
	cd testdata/protobuf/v1 && protoc --proto_path=. --descriptor_set_out=../../proto_old.pb --include_imports catalog.proto
	cd testdata/protobuf/v2 && protoc --proto_path=. --descriptor_set_out=../../proto_new.pb --include_imports catalog.proto
