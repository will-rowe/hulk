#!/usr/bin/env Rscript

####

library(grid)
library(pheatmap)

####

## Edit body of pheatmap:::draw_colnames, customizing it to your liking

## For pheatmap_1.0.8 and later:
draw_colnames_45 <- function (coln, gaps, ...) {
    coord = pheatmap:::find_coordinates(length(coln), gaps)
    x = coord$coord - 0.5 * coord$size
    res = textGrob(coln, x = x, y = unit(1, "npc") - unit(3,"bigpts"), vjust = 0.5, hjust = 1, rot = 45, gp = gpar(...))
    return(res)}

## 'Overwrite' default draw_colnames with your own version
assignInNamespace(x="draw_colnames", value="draw_colnames_45",ns=asNamespace("pheatmap"))

####

# Read in the csv
data <- read.csv("../out.csv")

# Label the rows
rownames(data) <- colnames(data)

# Transform for plotting
data <- as.matrix(data)

# Set up the image
png('hulk-smash-heatmap.png',width=30,height=30,units='in',res=500)

# Plot
pheatmap(data)
dev.off()
