<div align="center">
    <img src="paper/img/misc/hulk-logo-with-text.png?raw=true?" alt="hulk-logo" width="250">
    <h3><a style="color:#9900FF">H</a>istosketching <a style="color:#9900FF">U</a>sing <a style="color:#9900FF">L</a>ittle <a style="color:#9900FF">K</a>mers</h3>
    <hr/>
    <a href="https://travis-ci.org/will-rowe/hulk"><img src="https://travis-ci.org/will-rowe/hulk.svg?branch=master" alt="travis"></a>
    <a href='http://hulk.readthedocs.io/en/latest/?badge=latest'><img src='https://readthedocs.org/projects/hulk/badge/?version=latest' alt='Documentation Status' /></a>
    <a href="https://goreportcard.com/report/github.com/will-rowe/hulk"><img src="https://goreportcard.com/badge/github.com/will-rowe/hulk" alt="reportcard"></a>
    <a href="https://zenodo.org/badge/latestdoi/143890875"><img src="https://zenodo.org/badge/143890875.svg" alt="DOI"></a>
    <a href="https://github.com/will-rowe/hulk/blob/master/LICENSE"><img src="https://img.shields.io/badge/license-MIT-orange.svg" alt="License"></a>
    <a href="https://bioconda.github.io/recipes/hulk/README.html"><img src="https://anaconda.org/bioconda/hulk/badges/downloads.svg" alt="bioconda"></a>
    <a href="https://mybinder.org/v2/gh/will-rowe/hulk/master?filepath=paper%2Fanalysis-notebooks"><img src="https://mybinder.org/badge_logo.svg" alt="Binder"></a>
    <hr/>
</div>

> UPDATE: JULY 2019

> I no longer work for STFC. All versions of HULK pre 1.0.0 have been renamed and archived to the [STFC github](https://github.com/stfc/histogramSketcher). The STFC Hartree Centre are building genomic solutions based on these and other tools - if you are interested, please [contact them](hartree@stfc.ac.uk).

> This repo now hosts HULK >= version 1.0.0, which is a complete re-implementation of HULK and based solely off the method described in the [open-access paper](https://doi.org/10.1186/s40168-019-0653-2).

> I've tried to keep much of the syntax and existing functionality, but make sure to check the change log below. It's a work in progress but the master branch should be a close drop-in replacement for the old HULK (for sketching at least). There are a few algorithmic differences, mainly that HULK now uses **minimizers frequencies** for representing the underling microbiome sample.


> Importantly, this project is now **fully open source** and I can develop freely on it!

## Overview

`HULK` is a tool that creates small, fixed-size sketches from streaming microbiome sequencing data, enabling **rapid metagenomic dissimilarity analysis**. `HULK` approximates a [k-mer spectrum](https://bmcbioinformatics.biomedcentral.com/articles/10.1186/s12859-015-0875-7) from a FASTQ data stream, incrementally sketches it and makes similarity search queries against other microbiome sketches.

`HULK` works by collecting **minimizers** from sequences. Minimizers are assigned to a finite number of histogram bins using a [consistent jump hash](https://arxiv.org/abs/1406.2294); these bins are incremented as their corresponding minimizers are found. At set intervals (i.e. after X sequences have been processed), the bins are histosketched by `HULK`. Similarly to [MinHash sketches](https://en.wikipedia.org/wiki/MinHash), histosketches can be used to estimate similarity between sequence data sets.

The advantages of `HULK` include:

* it's fast and can run on a laptop
* **hulk sketches** are compact, fixed size and incorporate k-mer frequency information
* it works on data streams and does not require complete data instances
* it can use [concept drift](https://en.wikipedia.org/wiki/Concept_drift) for histosketching
* you get to type `hulk smash` into the command line...

Finally, you can use **hulk sketches** to with a Machine Learning classifier to predict microbiome sample origin (see [the paper](https://doi.org/10.1186/s40168-019-0653-2) and [BANNER](https://github.com/will-rowe/banner)).

## Change log

### version 1.0.1 (dev branch)

* WASM interface
  * run HULK locally and from a browser
  * based on my [baby-GROOT](https://github.com/will-rowe/baby-groot) user interface
* HULK will output additional sketches
  * KMV MinHash
  * HyperMinHash
* Indexing
  * re-implementation of the LSH Forest index

### version 1.0.0 (current release)

* fully re-written codebase
  * I've aimed for it to be largely backwards compatible with previous releases
* fully open-sourced!
  * MIT license ([OSI approved](https://opensource.org/licenses))
* algorithm changes
  * underlying histogram is now based on minimizer frequencies
  * count-min sketch for k-mer frequencies is now replaced with a fixed-size array and a jump-hash for minimizer placement
* changes to the `sketch` subcommand:
  * sketches saved to JSON by default (ala [sourmash](https://github.com/dib-lab/sourmash))
  * histosketch count-min sketch is no longer configurable by the user (this was Epsilon and Delta)
  * spectrum size is determined based on k-mer size
  * minCount for k-mer frequencies is removed
* changes to the `smash` subcommand:
  * operates on JSON input
  * outputs matrix as csv
* replaced some unecessary features
  * the functionality of the `print` and `distance` subcommands is available in the `smash` subcommand

### pre version 1.0.0

* all versions of HULK (and BANNER) pre v1.0.0 have been moved to the [UKRI github](https://github.com/stfc/histogramSketcher) and renamed. I can no longer work on these code bases.

## Installation

Check out the [releases](https://github.com/will-rowe/hulk/releases) to download a binary. Alternatively, install using Bioconda or compile the software from source.

### Bioconda

For versions <1.0.0, use bioconda. I will add the recipe for HULK 1.0.0 asap.

```bash
conda install -c bioconda hulk
```

### Source

`HULK` is written in Go (v1.12) - to compile from source you will first need the [Go tool chain](https://golang.org/doc/install). Once you have it, try something like this to compile:

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

`HULK` is called by typing **hulk**, followed by the subcommand you wish to run. There main subcommands are **sketch** and **smash**:

```bash
# Create a hulk sketch
gunzip -c microbiome.fq.gz | hulk sketch -o sketches/sampleA

#  Get a pairwise weighted Jaccard similarity matrix for a set of hulk histosketches
hulk smash -k 31 -m weightedjaccard -d ./sketches -o myOutfile
```

## Further Information & Citing

I'm working on some new documentation and this will be available on [readthedocs](http://hulk.readthedocs.io/en/latest/?badge=latest) soon.

A paper describing the `HULK` method is published in Microbiome:

>[Rowe WPM et al. Streaming histogram sketching for rapid microbiome analytics. Microbiome. 2019.](https://doi.org/10.1186/s40168-019-0653-2)
