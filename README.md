<div align="center">
    <img src="/paper/img/misc/hulk-logo-with-text.png?raw=true?" alt="hulk-logo" width="250">
    <h3><a style="color:#9900FF">H</a>istosketching <a style="color:#9900FF">U</a>sing <a style="color:#9900FF">L</a>ittle <a style="color:#9900FF">K</a>mers</h3>
    <hr>
    <a href="https://travis-ci.org/will-rowe/hulk"><img src="https://travis-ci.org/will-rowe/hulk.svg?branch=master" alt="travis"></a>
    <a href='http://hulk.readthedocs.io/en/latest/?badge=latest'><img src='https://readthedocs.org/projects/hulk/badge/?version=latest' alt='Documentation Status' /></a>
    <a href="https://goreportcard.com/report/github.com/will-rowe/hulk"><img src="https://goreportcard.com/badge/github.com/will-rowe/hulk" alt="reportcard"></a>
    <a href="https://github.com/will-rowe/hulk/blob/master/LICENSE"><img src="https://img.shields.io/badge/license-MIT-orange.svg" alt="License"></a>
    <a href="https://zenodo.org/badge/latestdoi/143890875"><img src="https://zenodo.org/badge/143890875.svg" alt="DOI"></a>
</div>

***

## Overview

`HULK` is a tool that creates small, fixed-size sketches from streaming microbiome sequencing data, enabling **rapid metagenomic dissimilarity analysis**. `HULK` generates a [k-mer spectrum](https://bmcbioinformatics.biomedcentral.com/articles/10.1186/s12859-015-0875-7) from a FASTQ data stream, incrementally sketches it and makes similarity search queries against other microbiome sketches.

It works by using [count-min sketching](https://en.wikipedia.org/wiki/Count%E2%80%93min_sketch) to create a k-mer spectrum from a data stream. After some reads have been added to a k-mer spectrum, `HULK` begins to process the counter frequencies and populates a [histosketch](https://exascale.info/assets/pdf/icdm2017_HistoSketch.pdf). Similarly to [MinHash sketches](https://en.wikipedia.org/wiki/MinHash), histosketches can be used to estimate similarity between microbiome samples.

The advantages of `HULK` include:

* it's fast and can run on a laptop in minutes
* **hulk sketches** are compact and a fixed size
* it works on data streams and does not require complete data instances
* it can use [concept drift](https://en.wikipedia.org/wiki/Concept_drift) for histosketching
* you get to type `hulk smash` into the command line...

Finally, you can use **hulk sketches** to with a Machine Learning classifier to bin microbiome samples (see [BANNER](https://github.com/will-rowe/banner)). More info on this coming soon...

## Installation

Check out the [releases](https://github.com/will-rowe/hulk/releases) to download a binary. Alternatively, install using Bioconda or compile the software from source.

### Bioconda

```
conda install hulk
```

> note: if using Conda make sure you have added the [Bioconda](https://bioconda.github.io/) channel first

### Source

`HULK` is written in Go (v1.9) - to compile from source you will first need the [Go tool chain](https://golang.org/doc/install). Once you have it, try something like this to compile:

```bash
# Clone this repository
git clone https://github.com/will-rowe/hulk.git

# Go into the repository and get the package dependencies
cd hulk
go get -d -t -v ./...

# Run the unit tests
go test -v ./...

# Compile the program
go build ./

# Call the program
./hulk --help
```

## Quick Start

`HULK` is called by typing **hulk**, followed by the subcommand you wish to run. There are three main subcommands: **sketch**, **distance** and **smash**. This quick start will show you how to get things running but it is recommended to follow the [documentation](http://hulk-documentation.readthedocs.io/en/latest/?badge=latest).

```bash
# Create a hulk sketch
gunzip -c microbiome.fq.gz | hulk sketch -p 8 -o sampleA

# Get similarity measures between two hulk sketches
hulk distance -1 sampleA.sketch -2 sampleB.sketch

#  Get a pairwise Jaccard Similarity matrix for a set of hulk sketches
hulk smash --jsMatrix -d ./dir-with-sketches-in -o my-jsMatrix

# Create a sketch matrix to train a Random Forest Classifier (see banner)
## smash all the sketches from one sample type (labeled 0)
hulk smash --bannerMatrix -o abx-treatedx -l 0
## smash all the sketches from another sample type (labeled 1), this time recursively
hulk smash  --bannerMatrix --sketchDir ./no-abx-sketches --recursive -o no-abx -l 1
# join both samples into one matrix
cat abx-treated.banner-matrix.csv no-abx.banner-matrix.csv > training.csv

# Train a Random Forest Classifier (make sure you have banner)
conda install banner
banner train --matrix training.csv

# Predict!
hulk sketch -f mystery-sample.fastq --stream -p 8 | banner predict -m banner.rfc
```

## Further Information & Citing

Please [readthedocs](http://hulk.readthedocs.io/en/latest/?badge=latest) for more extensive documentation and a [tutorial](https://hulk.readthedocs.io/en/latest/tutorial.html) will be forthcoming.

A preprint describing `HULK` is in preparation and I'll post a link soon... For now, here is the Genome Science 2018 abstract:

>[Rowe WPM et al. GROOT and HULK: Sketching microbiomes for resistome profiling and determining antibiotic dysbiosis. Genome Science (oral presentation) 2018.](/paper/genome-science-2018-abstract.md)
