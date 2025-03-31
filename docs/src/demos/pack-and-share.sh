#!/usr/bin/env bash

#######################################################
# Demo script to show how to pack and share a model between
# a data scientist and an app developer
#######################################################


. ../demo-magic.sh


########################
# Configure the options
########################
#
# speed at which to simulate typing. bigger num = faster
#
 TYPE_SPEED=25

#
# custom prompt
#
# see http://www.tldp.org/HOWTO/Bash-Prompt-HOWTO/bash-prompt-escape-sequences.html for escape sequences
#

# Reguires the kit CLI to be installed

## pull the example model we are going to use
EXAMPLE_MODEL="ghcr.io/kitops-ml/modelkit-examples/scikitlearn-tabular:latest"
NO_WAIT=true

rm -rf ~/.kitops
rm -rf model
rm -rf airsat
rm -rf deploy


mkdir model
kit unpack $EXAMPLE_MODEL -d ./model

DEMO_PROMPT="${GREEN}➜ ${CYAN}data-scientist \$ ${COLOR_RESET}"

# text color
# DEMO_CMD_COLOR=$BLACK

# hide the evidence
clear
wait

# show the contents of the model folder
pe "tree ./model"

PROMPT_TIMEOUT=3
wait

# pe "cat ./model/Kitfile"
# PROMPT_TIMEOUT=3
# wait

# pack the model into a modelkit
pe "cd model"
p " ### Packing the model, dataset, and Jupyter notebook into a ModelKit"
pe "kit pack . -t airsat:integrate"
pe "kit list"

## tag for remote
p " ### Tagging the ModelKit with remote reference"
pe "kit tag airsat:integrate ghcr.io/kitops-ml/demos/airsat:integrate"
pe "kit list"

# push the model to the remote
p " ### Pushing the ModelKit to the remote"
pe "kit push ghcr.io/kitops-ml/demos/airsat:integrate"

PROMPT_TIMEOUT=3
wait


### SWITCH TO APP DEVELOPER ###
DEMO_PROMPT="${GREEN}➜ ${CYAN}app-developer \$ ${COLOR_RESET}"
cd ..
clear

## pull the model
p " ### Pulling the ModelKit from the remote to integrate into the application"
pe "kit pull ghcr.io/kitops-ml/demos/airsat:integrate"

pe "mkdir airsat"

## unpack only the model and datasets
p " ### Unpacking only model into a directory"
pe "kit unpack ghcr.io/kitops-ml/demos/airsat:integrate --model -d ./airsat"
pe "tree ./airsat"

## tag for deployment and push
p " ### Integration is done tagging the ModelKit for deployment and pushing to the remote"
pe "kit tag ghcr.io/kitops-ml/demos/airsat:integrate ghcr.io/kitops-ml/demos/airsat:deploy"
pe "kit push ghcr.io/kitops-ml/demos/airsat:deploy"


# show a prompt so as not to reveal our true nature after
# the demo has concluded
cmd
p ""
exit
