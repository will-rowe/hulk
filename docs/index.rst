Welcome to HULK's wiki!
============================

`HULK` is a tool that creates small, fixed-size sketches from streaming microbiome sequencing data, enabling **rapid metagenomic dissimilarity analysis**. `HULK` generates a k-mer spectrum_ from a FASTQ data stream, incrementally sketches it and makes similarity search queries against other microbiome sketches.

It works by using count-min-sketching_ to create a k-mer spectrum from a data stream. After some reads have been added to a k-mer spectrum, `HULK` begins to process the counter frequencies and populates a histosketch_. Similarly to MinHash_, histosketches can be used to estimate similarity between microbiome samples.

The advantages of `HULK` include:

* it's fast and can run on a laptop in minutes
* **hulk sketches** are compact and a fixed size
* it works on data streams and does not require complete data instances
* it can use concept-drift_ for histosketching
* you get to type `hulk smash` into the command line...

Finally, you can use **hulk sketches** to with a Machine Learning classifier to bin microbiome samples (see BANNER_).

----

Contents
------------------------------------------------------
.. toctree::
   :maxdepth: 2

   using-hulk
   using-banner
   tutorial-clustering
   tutorial-indexing

----

Reference
------------------------------------------------------

Check out the preprint_ for more information.

----

|
------------------------------------------------------

.. _spectrum: https://doi.org/10.1186/s12859-015-0875-7
.. _histosketch: https://exascale.info/assets/pdf/icdm2017_HistoSketch.pdf
.. _MinHash: https://en.wikipedia.org/wiki/MinHash
.. _count-min-sketching: https://en.wikipedia.org/wiki/Count%E2%80%93min_sketch
.. _concept-drift: https://en.wikipedia.org/wiki/Concept_drift
.. _BANNER: https://github.com/will-rowe/banner
.. _preprint: https://doi.org/10.1101/408070
