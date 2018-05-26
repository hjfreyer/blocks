
import argparse
#import tensorflow as tf
import numpy as np

import snake

MAX_ROUNDS = 1000
SIZE = 11

UNITS = 16
INPUTS = 9
OUTPUTS = 3
LAYERS = 3

MODEL_LEN = UNITS*UNITS*LAYERS + UNITS*INPUTS + UNITS*OUTPUTS


POP_SIZE = 10
POP_KEEP = 0.1

def apply_model(g, model):
	model = model.reshape((UNITS*LAYERS + INPUTS + OUTPUTS, UNITS))
	
	layer1 = model[:INPUTS, :].transpose()
	last_layer = model[INPUTS:INPUTS+OUTPUTS, :]
	layers = model[INPUTS+OUTPUTS:, :].reshape((LAYERS, UNITS, UNITS))

	v = np.array(snake.get_stimuli(g)).transpose()
	v = np.matmul(layer1, v)
	v = np.tanh(v)
	for layer in layers:
		v = np.matmul(layer, v)
		v = np.tanh(v)
	output = np.matmul(last_layer, v)
	return 'LCR'[output.argmax()]
	
	

def do_game(model):
	g = snake.InitGame(SIZE)
	
	step = 0
	while step < 1000 and g.state == 'LIVE':
		move = apply_model(g, model)
		g = snake.AdvanceGame(g, move)
		step += 1
	
	return -len(g.body)

parser = argparse.ArgumentParser()

def main(argv):
	args = parser.parse_args(argv[1:])
	
	population = np.random.rand(POP_SIZE, MODEL_LEN)
	results = [do_game(p) for p in population]
	print results
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
	import cProfile
	cProfile.run('main(sys.argv)', sort='cumtime')