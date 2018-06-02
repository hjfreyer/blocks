
import argparse
#import tensorflow as tf
import numpy as np
import random

import snake

MAX_ROUNDS = 1000
SIZE = 11

UNITS = 12
INPUTS = 9
OUTPUTS = 3
LAYERS = 1

MODEL_LEN = UNITS*(INPUTS+1) + (UNITS+1)*OUTPUTS
MUTATION_RATE = 0.1

POP_SIZE = 50
POP_KEEP = 0.3

PLAYS = 5

def apply_model(g, model):
	layer1, layer2 = model[:(INPUTS+1) * UNITS], model[(INPUTS+1)*UNITS:]
	layer1 = layer1.reshape((UNITS, INPUTS + 1))

	layer2 = layer2.reshape((OUTPUTS, UNITS+1))
	
	v = np.array(snake.get_stimuli(g) + [1]).transpose()
	v = np.tanh(np.matmul(layer1, v))
	v = np.concatenate((v, [1]))

	v = np.matmul(layer2, v)	
	return 'LCR'[v.argmax()]
	
	

def do_game(model):
	g = snake.InitGame(SIZE)
	
	step = 0
	while step < 1000 and g.state == 'LIVE':
		move = apply_model(g, model)
		g = snake.AdvanceGame(g, move)
		step += 1
	
	return -len(g.body)
	
def eval_model(model):
	return np.mean([do_game(model) for i in range(PLAYS)])

parser = argparse.ArgumentParser()

def crossover(a, b):	
	cutoff = random.randint(0, MODEL_LEN)
	return np.concatenate([a[:cutoff], b[cutoff:]])

def do_generation(results):
	population = results[np.argsort(results[:, 1]), 0]
	population = population[:int(POP_SIZE*POP_KEEP)]

	children = list(population)
	while len(children) < POP_SIZE:
		child = crossover(*random.sample(population, 2))
		child += np.where(np.random.uniform(size=len(child)) < MUTATION_RATE, np.random.normal(scale=5, size=len(child)), 0)
		children.append(child)			
	
	results = np.array([[p, eval_model(p)] for p in population])
	return results

def main(argv):
	args = parser.parse_args(argv[1:])
	
	population = np.random.normal(size=(POP_SIZE, MODEL_LEN))
	results = np.array([[p, eval_model(p)] for p in population])
	idx = 1
	while True:
		print 'Generation', idx
		results = do_generation(results)
		print 'Min =', results[:, 1].min(), "Mean: ", np.mean(results[:, 1])
		idx += 1

"""
	model = tf.get_variable("model", shape=(MODEL_LEN))
	y = tf.py_func(do_game, [model], tf.float32)
	
	optimizer = tf.train.AdamOptimizer()
	train = optimizer.minimize(y)

	with tf.Session() as sess:
		sess.run(tf.global_variables_initializer())
		
		for step in range(1):
			sess.run(train)
			print sess.run(x),sess.run(y)
"""

if __name__ == '__main__':
	import sys
	#import cProfile
	main(sys.argv)
	#cProfile.run('main(sys.argv)', sort='cumtime')