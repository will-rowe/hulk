#!/usr/bin/env bash

# This script automates the submission of Hulk jobs to Scafell.
# It will run hulk on any *fq.gz files it finds in the current directory, storing all sketches to a new subdir (based on the job parameters)
# A sketch will be made for each file; all sketches can then be combined into a pairwise similarity matrix.

# PARAMETERS
## general
CPU=12
QUEUE=scafellpikeSKL
WALL=0:20
## hulk
K=11
S=256

# JOB SUBMISSION
mkdir hulk-sketches-k${K}-s${S} && cd $_
for i in ../*.gz
do
outfile=${i##*/}
CMD="gunzip -c $i | hulk sketch -p ${CPU} -s ${S} -k ${K} -o ${outfile%%.fq.gz}"
echo $CMD | bsub -n ${CPU} -R "span[ptile=${CPU}]" -W ${WALL} -q ${QUEUE}
done
