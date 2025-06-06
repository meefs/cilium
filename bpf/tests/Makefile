# Copyright Authors of Cilium
# SPDX-License-Identifier: (GPL-2.0-only OR BSD-2-Clause)

include ../../Makefile.defs

MAKEFLAGS += -r

CLANG ?= clang

FLAGS := -I$(ROOT_DIR)/bpf -I$(ROOT_DIR)/bpf/include -g

# Base CFLAGS (CLANG_FLAGS) are read from pkg/datapath/loader/compile.go
# to keep them synced with the loader
CLANG_FLAGS := $(FLAGS) $(shell $(GO) run $(ROOT_DIR)/pkg/datapath/loader/tools/clang_cflags.go)

# Create dependency files for each .o file.
CLANG_FLAGS += -MD

# Mimics the mcpu values set by cilium-agent. See GetBPFCPU().
ifneq ($(KERNEL),54)
CLANG_FLAGS += -mcpu=v3
else
CLANG_FLAGS += -mcpu=v2
endif

.PHONY: all clean run

TEST_OBJECTS = $(patsubst %.c, %.o, $(wildcard *.c))

%.o: %.c $(LIB)
	@$(ECHO_CC)
	@# Remove the .o file to force recompilation, only rely on make's caching, not clangs
	rm -f $@
	$(QUIET) ${CLANG} ${CLANG_FLAGS} -c $< -o $@

%.ll: %.c $(LIB)
	@$(ECHO_CC)
	$(QUIET) ${CLANG} ${CLANG_FLAGS} -c -emit-llvm $< -o $@

%.i: %.c $(LIB)
	@$(ECHO_CC)
	$(QUIET) ${CLANG} ${CLANG_FLAGS} -E -c $< -o $@

all: $(TEST_OBJECTS)

clean:
	rm -f $(wildcard *.ll)
	rm -f $(wildcard *.o)
	rm -f $(wildcard *.i)
	rm -f $(wildcard *.d)

BPF_TEST_FLAGS:= $(GO_TEST_FLAGS)
ifneq ($(shell id -u), 0)
		BPF_TEST_FLAGS += -exec "sudo -E"
endif
ifeq ($(V),1)
    BPF_TEST_FLAGS += -test.v
endif
ifeq ($(COVER),1)
	ifndef COVERFORMAT
		COVERFORMAT:=html
	endif
    BPF_TEST_FLAGS += -coverage-report $(ROOT_DIR)/bpf-coverage.$(COVERFORMAT) -coverage-format $(COVERFORMAT)
ifdef NOCOVER
    BPF_TEST_FLAGS += -no-test-coverage "$(NOCOVER)"
endif
endif
ifeq ($(INSTRLOG),1)
    BPF_TEST_FLAGS += -instrumentation-log $(ROOT_DIR)/test/bpf-instrumentation.log
endif
ifdef RUN
    BPF_TEST_FLAGS += -run $(RUN)
endif
ifdef BPF_TEST_DUMP_CTX
    BPF_TEST_FLAGS += -dump-ctx
endif
ifdef BPF_TEST_FILE
	BPF_TEST_FLAGS += -test $(BPF_TEST_FILE)
endif

run: $(TEST_OBJECTS)
	$(QUIET)$(GO) test ./bpftest \
		-bpf-test-path $(ROOT_DIR)/bpf/tests \
		$(BPF_TEST_FLAGS) \
	| $(GOTEST_FORMATTER)

-include $(TEST_OBJECTS:.o=.d)
