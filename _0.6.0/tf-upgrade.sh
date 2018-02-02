#!/bin/bash
# 4.4 bash version required

find . -path ./.terraform -prune -o -type f -name '*.tf*' \
  -exec sh -c \
  "sed -i.bak -E 's/terraform_remote_state\.([^\.]+)\.output/terraform_remote_state.\1/' {} && rm -f {}.bak" \;

# Shared services remote state definition was a symlink
rm -f shared-services.tf.tpl
cp ../shared-services.tf ./shared-services.tf.tpl
rm -f version/shared-services.tf.tpl
cp ../shared-services.tf ./version/shared-services.tf.tpl

# Data sources
find . -path ./.terraform -prune -o -type f -name '*.tf*' \
  -exec sh -c \
  "sed -i.bak -E 's/resource \"terraform_remote_state\" /data \"terraform_remote_state\" /' {} && rm -f {}.bak" \;
find . -path ./.terraform -prune -o -type f -name '*.tf*' \
  -exec sh -c \
  "sed -i.bak -E 's/terraform_remote_state\./data\.terraform_remote_state\./' {} && rm -f {}.bak" \;

find . -path ./.terraform -prune -o -type f -name '*.tf*' \
  -exec sh -c \
  "sed -i.bak -E 's/resource \"template_file\" /data \"template_file\" /' {} && rm -f {}.bak" \;
find . -path ./.terraform -prune -o -type f -name '*.tf*' \
  -exec sh -c \
  "sed -i.bak -E 's/template_file\./data\.template_file\./' {} && rm -f {}.bak" \;
