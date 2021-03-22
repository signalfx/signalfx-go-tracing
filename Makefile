# ALL_MODULES includes ./* dirs (excludes . dir)
ALL_MODULES := $(shell find . -type f -name "go.mod" -exec dirname {} \; | sort )

GO = go

.PHONY: mod-tidy
mod-tidy:
	@set -e; for dir in $(ALL_MODULES); do \
	  echo "$(GO) mod tidy in $${dir}"; \
	  (cd "$${dir}" && $(GO) mod tidy); \
	done

DEPENDABOT_PATH=./.github/dependabot.yml
DEPENDABOT_MODULES=$(filter-out ".", $(ALL_MODULES))
.PHONY: gendependabot
gendependabot:
	@echo "Recreate dependabot.yml file"
	@echo "# File generated by \"make gendependabot\"; DO NOT EDIT.\n" > ${DEPENDABOT_PATH}
	@echo "version: 2" >> ${DEPENDABOT_PATH}
	@echo "updates:" >> ${DEPENDABOT_PATH}
	@echo "Add entry for \"/\""
	@echo "  - package-ecosystem: \"gomod\"\n    directory: \"/\"\n    schedule:\n      interval: \"weekly\"" >> ${DEPENDABOT_PATH}
	@set -e; for dir in $(DEPENDABOT_MODULES); do \
		(echo "Add entry for \"$${dir:1}\"" && \
		  echo "  - package-ecosystem: \"gomod\"\n    directory: \"$${dir:1}\"\n    schedule:\n      interval: \"weekly\"" >> ${DEPENDABOT_PATH} ); \
	done

