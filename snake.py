
import collections
import json
import random

import numpy as np

SnakeGame = collections.namedtuple('SnakeGame', 'size state fruit body direction')

MOVES = list('LRC')

def OOB(size, pt):
	return (pt < 0).any() or (size <= pt).any()

def DirPlusMove(dir, move):
	if move == 'C': return dir
	if move == 'L':
		return np.array([dir[1], -dir[0]])
	if move == 'R':
		return np.array([-dir[1], dir[0]])


def RandomPoint(size):
	return np.random.random_integers(0, size-1, 2)

def InitGame(size):
	head = np.array([size / 2, size / 2])
	dir = np.array([0, -1])
	return SnakeGame(size=size,
					 state='LIVE',
					 fruit=RandomPoint(size),
					 body=np.array([head, head-dir]),
					 direction=dir)

def AdvanceGame(g, move):
	new_dir = DirPlusMove(g.direction, move)
	new_head = g.body[0] + new_dir
	
	if OOB(g.size, new_head):
		return g._replace(state='DEAD')
	
	new_body = np.concatenate([[new_head], g.body])
	if (new_head == g.fruit).all():	
		new_fruit = RandomPoint(g.size)
		while (new_fruit == new_body).all(axis=1).any():
			new_fruit = RandomPoint(g.size)
	else:
		new_fruit = g.fruit
		new_body = new_body[:-1, :]

	if (new_head == new_body[1:, ]).all(axis=1).any():
		return g._replace(state='DEAD')
	return g._replace(body=new_body, fruit=new_fruit, direction=new_dir)

def ExportGame(stream, games):
	data = dict(width=games[0].size, height=games[0].size, steps=[])
	for g in games:	
		step = dict(pts=[], comment='')
		
		fx, fy = g.fruit
		pts = [dict(x=fx, y=fy, type='fruit')]
		pts += [dict(x=x, y=y, type='snake')
				for (x, y) in g.body]
		data['steps'].append(dict(pts=pts, comment=str(get_stimuli(g))))
		
 	json.dump(data, stream, indent=2)
 	

def get_stimuli(g):
	l_eye = get_eye(g, DirPlusMove(g.direction, 'L'))
	c_eye = get_eye(g, DirPlusMove(g.direction, 'C'))
	r_eye = get_eye(g, DirPlusMove(g.direction, 'R'))
	return l_eye + c_eye + r_eye

def vect_div(p, q):
	codirectional = (np.sign(p) == np.sign(q)).all(axis=1)
	return np.where(codirectional, np.abs(np.sum(p, axis=1)), 0)

def get_eye(g, dir):
	head = g.body[0]
	
	# Walls. Do it dumbly.
	if dir[0] == 1:  # Right.
		wall_idx = g.size - head[0]
	elif dir[0] == -1:  # Left.
		wall_idx = head[0] + 1
	elif dir[1] == 1:  # Down.
		wall_idx = g.size - head[1]
	elif dir[1] == -1:  # Up.
		wall_idx = head[1] + 1
	else:
		raise Exception("Ahhhh")
		
	fruit_div = vect_div([g.fruit-head], dir)
	fruit_idx = int(fruit_div[0]) if fruit_div[0] > 0 else 100
	
	body_div = vect_div(g.body[1:] - head, dir)
	body_div = body_div[0 < body_div]	
	body_idx = int(body_div.min()) if len(body_div) else 100
	return [wall_idx, fruit_idx, body_idx]

 	
if __name__ == '__main__':
	g = InitGame(11)
	games = [g]

	tries = 0
	while len(games[-1].body) < 5:
		tries += 1
		g = InitGame(11)
		games = [g]
		while g.state != 'DEAD':
			if len(games) > 100:
				break
			g = AdvanceGame(g, random.choice(MOVES))
			games.append(g)
	print 'tries = ', tries
	with open('ui/data.json', 'wb') as f:
		ExportGame(f, games)