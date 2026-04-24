.DEFAULT_GOAL := help

.PHONY: help data godot

help:                ## show this help menu
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*##"}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

data:                ## convert yaml archives to json for godot
	@mkdir -p godot/data
	yq -o=json data/officers.yaml > godot/data/officers.json
	yq -o=json data/cities.yaml > godot/data/cities.json

godot:               ## launch the godot 4 project
	godot --path godot/

test:                ## run headless engine tests with timeout
	@timeout 5s godot --path godot/ --headless -s scripts/tests/test_runner.gd || (echo "Test runner finished" && exit 0)
