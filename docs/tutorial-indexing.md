
# Tutorial: part 2

## indexing microbiome samples

> Note: This tutorial uses the latest release of HULK (0.1.0).

***

So, in part one we started to look at a basic use case for HULK:

* We have a bunch of existing microbiome samples and we want to compare some new samples against these, pulling out similar samples from our collection and then do some cool science with them.

We looked at how HULK histosketches can be used to cluster microbiome samples by body site. In this part of the tutorial, we now are going to index these histosketches and see how we can then query this index and pull out similar microbiome samples.


### Generate and search the index

* First of all, make a sub directory for each body site and split our histosketches by body site. This is just to make things a bit simpler in terms of querying the index later.

```
cd histosketches-k11-m1-c10-s256
mkdir airways gastrointestinal_tract oral skin urogenital_tract
mv Air* ./airways/ && mv Gastro* ./gastro*/ && mv Oral* ./oral/ && mv Skin* ./skin/ && mv Uro* ./uro*/
rm ./*/*.log
```

* Separate some query histosketches from our collection of histosketches

```
cd ..
mkdir cami-queries
for i in histosketches-k11-m1-c10-s256/*; do mkdir cami-queries/${i##*/}; ls $i | sort -R | tail -n 1 | xargs -I {} sh -c "mv $i/{} cami-queries/${i##*/}"; done
```

* Create an index of the remaining CAMI histosketches, using a Jaccard similarity theshold of 65%

```
hulk index -r create -n cami.index -j 0.65 -d ./histosketches-k11-m1-c10-s256/ --recursive
```

* Search the index using our queries

```
hulk index -r search -n cami.index -j 0.65 -d cami-queries/ --recursive > hulk-search-results.txt
```

### Interpretation

So we randomly took a histosketch from each body site and put this to one side to use as our search queries. The remaining histosketches we indexed using `hulk index`, which is a hulk subcommand to create a search index using the [Locality Sensitive Hashing Forest](http://ilpubs.stanford.edu:8090/678/1/2005-14.pdf) scheme.

You might get different results to me as we used a random set of query histosketches. But, in my case I got the following:

```
query:	cami-queries/airways/Airways.20.sketch
hit 0:	histosketches-k11-m1-c10-s256/airways/Airways.17.sketch
hit 1:	histosketches-k11-m1-c10-s256/oral/Oral.17.sketch
hit 2:	histosketches-k11-m1-c10-s256/airways/Airways.13.sketch
hit 3:	histosketches-k11-m1-c10-s256/airways/Airways.18.sketch
----
query:	cami-queries/gastrointestinal_tract/Gastrointestinal_tract.12.sketch
hit 0:	histosketches-k11-m1-c10-s256/gastrointestinal_tract/Gastrointestinal_tract.11.sketch
hit 1:	histosketches-k11-m1-c10-s256/gastrointestinal_tract/Gastrointestinal_tract.17.sketch
----
query:	cami-queries/oral/Oral.11.sketch
----
query:	cami-queries/skin/Skin.11.sketch
hit 0:	histosketches-k11-m1-c10-s256/skin/Skin.19.sketch
----
query:	cami-queries/urogenital_tract/Urogenital_tract.18.sketch
----
```

Again, not bad. Airways, GI_tract and Skin queries all returned other microbiome samples from the same body site (plus an extra Oral sample for the Airways query). The Oral and UG_tract didn't return any similar samples at this similarity threshold (65% Jaccard similarity), which is probably because we randomly selected those few pesky samples from the previous part of the tutorial that didn't cluster very well. If we dropped the similarity threshold, we will get more search results but also some more false positives (microbiome samples from other body sites) - which isn't the end of the world as we are trying to find similar samples to work with.

### Next steps

I'll add more to the tutorial soon, taking this indexing idea a bit further and then I'll add an extra part to the tutorial to cover the Random Forest Classifier I used in the paper.
