#!/usr/bin/env python
####
# Imports
####
import os
import sys
import time
import argparse
import numpy as np
import pickle
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import accuracy_score, f1_score, precision_score, recall_score, classification_report, confusion_matrix
try:
	from version import version_number
except:
	version_number = 'version unknown'

####
# Classes
####
class Banner(object):
	def __init__(self):
		parser = argparse.ArgumentParser(
			description='banner does this',
			usage='''banner <subcommand> [<args>]

subcommands:
   train	Train does this...
   predict	Predict does this...
''')
		parser.add_argument('subcommand', help='Subcommand to run')
		# parse_args defaults to [1:] for args, but you need to
		# exclude the rest of the args too, or validation will fail
		args = parser.parse_args(sys.argv[1:2])
		if not hasattr(self, args.subcommand):
			print ("Unrecognized subcommand")
			parser.print_help()
			exit(1)
		# use dispatch pattern to invoke method with same name
		getattr(self, args.subcommand)()

	"""
	the train method takes a matrix of sketches, with the final column containing the labels (ints) and trains a RFC
	"""
	def train(self):
		parser = argparse.ArgumentParser(
			description='Train takes a banner-matrix file from hulk and uses it to train a Random Forest Classifier')
		# this checks the file and returns an open handle
		parser.add_argument('-m', '--matrix', required=True, type=lambda x: fileCheck(parser, x), help='The matrix from hulk smash')
		parser.add_argument('-o', '--outFile', required=False, default='banner.rfc', help='Where to write the model to')
		args = parser.parse_args(sys.argv[2:])
		print ("##########\n# BANNER #\n##########")
		print ("running train")
		print('loading sketch matrix: {}' .format(args.matrix.name))
		# load the data
		data = np.loadtxt(fname = args.matrix, delimiter = ',', dtype = int)
		args.matrix.close()
		# split to features and labels
		features = data[:, :-1]
		labels = data[:, -1]
		print("training...")
		# split to training and testing (0.20 = 20% for testing) TODO: allow ratio to be set by user
		trainFeatures, trainLabels, testFeatures, testLabels = splitData(features, labels, 0.20)
		# create a classifier TODO: allow customisation here too...
		bannerRFC=RandomForestClassifier(bootstrap=True, class_weight=None, n_estimators=100, n_jobs=1)
		# train
		bannerRFC.fit(trainFeatures,trainLabels)
		trainPredictions = bannerRFC.predict(trainFeatures)
		print("training accuracy: {}" .format(accuracy_score(trainLabels, trainPredictions)))
		print("testing...")
		# test
		testPredictions = bannerRFC.predict(testFeatures)
		print("f1 score: {}" .format(f1_score(testLabels, testPredictions, average="macro")))
		print("precision: {}" .format(precision_score(testLabels, testPredictions, average="macro")))
		print("recall score: {}" .format(recall_score(testLabels, testPredictions, average="macro")))
		print("test accuracy: {}" .format(accuracy_score(testLabels, testPredictions)))
		# save the model to disk
		pickle.dump(bannerRFC, open(args.outFile, 'wb'))
		print("saved model to disk: {}" .format(args.outFile))
		print("finished.")

	"""
	the predict method collects sketches from STDIN and classifies them using the RFC model
	"""
	def predict(self):
		parser = argparse.ArgumentParser(
			description='Predict collects sketches from STDIN and classifies them using a RFC')
		parser.add_argument('-m', '--model', required=True, help='The model that banner trained')
		parser.add_argument('-p', '--probability', required=False, default=0.90, type=float, help='The probability threshold for reporting classifications')
		parser.add_argument('-v', '--verbose', required=False, default=False, help='Print all predictions and probability, even if threshold not met yet')
		args = parser.parse_args(sys.argv[2:])
		print ("##########\n# BANNER #\n##########")
		print ("running predict")
		print("loading model: {}" .format(args.model))
		print("waiting for sketches...")
		# load the model
		bannerRFC = pickle.load(open(args.model, 'rb'))
		# wait for input from STDIN
		for line in sys.stdin:
			query = np.fromstring(line, sep = ',', dtype = int)
			# get the query into format
			query = np.asarray(query)
			query = np.reshape(query, (1, -1))
			# classify the query
			prediction = bannerRFC.predict(query)
			probability = bannerRFC.predict_proba(query)
			if args.verbose:
				print("predicted: {}" .format(prediction[0]))
				print("probability: {} {}" .format(probability[0][0], probability[0][1]))
			# exit if threshold met
			if (probability[0][0] >= args.probability) | (probability[0][1] >= args.probability):
				print("probability threshold met!")
				print("predicted: {}" .format(prediction[0]))
				print("probability: {} {}" .format(probability[0][0], probability[0][1]))
				print("finished.")
				sys.exit(0)
			# if stdin finished but we're still here, not prediction could be made with that probability threshold
			print("could not make prediction within probability threshold!")
			print("finished.")
			sys.exit(0)

####
# Functions
####
# splitData splits the input data into training and testing sets
def splitData(features, labels, test_size):
	total_test_size = int(len(features) * test_size)
	np.random.seed(2)
	indices = np.random.permutation(len(features))
	train_features = features[indices[:-total_test_size]]
	train_labels = labels[indices[:-total_test_size]]
	test_features  = features[indices[-total_test_size:]]
	test_labels  = labels[indices[-total_test_size:]]
	return train_features, train_labels, test_features, test_labels

# fileCheck makes sure the file exists
def fileCheck(parser, arg):
	if not os.path.exists(arg):
		parser.error("The file %s does not exist!" % arg)
	else:
		return open(arg, 'r')  # return an open file handle

####
# Main
####
if __name__ == '__main__':
	try:
		Banner()
	except KeyboardInterrupt:
		print("\ncanceled by user!")
		sys.exit(0)
