.PHONY: lint format format-check

lint:
	luacheck lua/ plugin/

format:
	stylua lua/ plugin/

format-check:
	stylua --check lua/ plugin/
