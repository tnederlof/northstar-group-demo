# Makefile (repo root)
.PHONY: %
%:
	@$(MAKE) -C demo $@
