# Using HULK

**HULK** uses a subcommand syntax; call the main program with `hulk` and follow it by the subcommand to indicate what  action to take.

This page will cover a worked example, details on the available HULK subcommands and some tips for using the program.

***

## An example

This example will take two microbiome samples, histosketch them and then get the weighted Jaccard Similarity of the two samples

Get some sequence data:
```
fastq-dump SRR4454612
fastq-dump SRR4454613
```

Histosketch the samples
```
hulk sketch -f SRR4454612.fastq -o sampleA
hulk sketch -f SRR4454613.fastq -o sampleB
```

Get similarity measures between two hulk sketches
```
hulk distance -1 sampleA.sketch -2 sampleB.sketch
```

***

## HULK subcommands

### sketch

The ``sketch`` subcommand is used to generate a histosketch from a stream of sequence reads. Here is an example:

```
gunzip -c microbiome.fastq.gz | hulk sketch -p 8 -k 11 -s 256 -o sampleA
```

The above command will stream a set of reads into HULK, which will then histosketch the reads.

Flags explained:

* ``-p``: how many processors to use for histosketching
* ``-k``: the k-mer size to use
* ``-s``: the number of elements to keep in the histosketch (i.e. the sketch length)
* ``-o``: the basename for the output files

You don't have to histosketch only one file, or even a whole data stream. The histosketching will stop when no more data is received, or if it has been told to stop.

Some more flags that can be used:

* ``-f``: this flag tells HULK to stream from a specified file (or list of files)
* ``-c``: maximum memory (MB) used by each CMS to store counts (higher MB = slower yet more accurate k-mer counting)
* ``-m``: minimum count number for a kmer bin to be considered for histosketch incorporation
* ``-x``: the decay ratio used for concept drift (1.00 = concept drift disabled)
* ``-i``: size of read sampling interval (0 == no interval)
* ``--streaming``: writes the sketches to STDOUT (as well as to disk)
* ``--fasta``: tells HULK that the input file(s) is in FASTA format

### distance

The ``distance`` subcommand is used to compare two histosketches. Here is an example:

```
hulk distance -1 a.sketch -2 b.sketch -m weightedjaccard
```

Flags explained:

* ``-1``: the first histosketch to compare
* ``-2``: the second histosketch to compare
* ``-m``: the distance metric to use (braycurtis/canberra/euclidean/jaccard/weightedjaccard)

### smash

The ``smash`` subcommand is used to gather multiple histoskeches. It is for 2 purposes, either to perform pairwise distance comparison between 2 or more samples, or to prepare a set of histosketches as vectors for machine learning. Here is an example of the former:

```
hulk smash --wjsMatrix -o mySamples
```

Flags explained:

* ``--wjsMatrix``: what matix to smash the histosketches into (here, pairwise weighted Jaccard similarity)
* ``-o``: the basename of the output file(s)

Some more flags that can be used:

* ``--jsMatrix``: output pairwise Jaccard similarity comparison matrix
* ``--bannerMatrix``: output histosketch matrix for BANNER (or other machine learning)
* ``-d``: which directory to collect the histosketches for smashing
* ``--recursive``: also look for histosketches in the sub-directories from ``-d``

The ``smash`` subcommand can give multiple outputs, i.e you can give it both ``--wjsMatrix`` and ``--bannerMatrix`` and it will output two files.

### index

The ``index`` subcommand is used to create, append and search an LSH Forest index of histosketches. Here is an example of creating and searching an index:

```
hulk index -r create -n cami.index -j 0.65 -d ./histosketches-k11/ --recursive

hulk index -r search -n cami.index -j 0.65 -d ./histosketch-queries/ --recursive
```

Flags explained:

* ``-r``: specifies the indexing function (create/add/search)
* ``-n``: the file name for the index
* ``-j``: the Jaccard similarity threshold for storing and retrieving histosketches
* ``-d``: the directory containing the histosketches for indexing/querying
* ``--recursive``: also look for histosketches in the sub-directories from ``-d``
