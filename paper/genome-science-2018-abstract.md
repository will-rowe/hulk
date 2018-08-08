#GROOT and HULK: Sketching microbiomes for resistome profiling and determining antibiotic dysbiosis

Will P. M. Rowe<sup>1&</sup>, Anna P. Carrieri<sup>2</sup>, Edward O. Pyzer-Knapp<sup>2</sup>, Lindsay J. Hall<sup>3</sup>, Martyn D. Winn<sup>1</sup>

<sup>1</sup> Scientific Computing Department, STFC Daresbury Laboratory, UK

<sup>2</sup> IBM Research, The Hartree Centre, UK

<sup>3</sup> Quadram Institute Bioscience, Norwich Research Park, Norwich, UK

<sup>&</sup> will.rowe@stfc.ac.uk

## Abstract

###Motivation

Antimicrobial resistance (AMR) remains a major threat to global health. Profiling the collective AMR genes within a microbiome (the ‘resistome’) and determining antibiotic dysbiosis facilitates greater understanding of AMR gene diversity and dynamics; allowing for gene surveillance, individualized treatment of bacterial infections and more sustainable use of antimicrobials. However, these analyses can be complicated by high similarity between reference genes, as well as the sheer volume of sequencing data and the complexity of analysis workflows.

We have developed efficient and accurate methods for resistome profiling and determining antibiotic dysbiosis that address these complications and improve upon currently available tools.

###Results

We present GROOT and HULK, two methods that utilise data sketching for rapid microbiome comparisons and similarity-search queries that can be performed in real-time on sequence data streams.

* GROOT

GROOT combines variation graph representation of gene sets with a locality-sensitive hashing forest indexing scheme to allow for fast classification and alignment of metagenomic sequence reads to known AMR gene variants.
On a set of clinical preterm infant microbiome samples, we show that GROOT can generate a resistome profile in 2 minutes using a single CPU (per sample), is more accurate than existing tools and can identify acquisition of AMR gene variants over time (e.g. gain of extended spectrum beta lactamase activity).

* HULK

HULK employs streaming histogram sketching of k-mer spectra to obtain a sample signature that is suitable for similarity testing and machine learning classifiers.
We show that HULK can sketch these same preterm infant microbiome samples in an equivalent time and can differentiate between antibiotic and non-antibiotic treated samples, enabling blinded clinical samples to be classified by antibiotic treatment history using a gaussian process classifier (accuracy 0.953, F1-score 0.967, Precision 0.972).

###Availability and implementation
GROOT and HULK are  written in Go and available at [https://github.com/will-rowe/groot](https://github.com/will-rowe/groot) and [https://github.com/will-rowe/hulk](https://github.com/will-rowe/hulk) (MIT licenses).
