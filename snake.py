
import collections
import json
import random

import numpy as np

Point = collections.namedtuple('Point', 'x y')
SnakeGame = collections.namedtuple('SnakeGame', 'width height state fruit head body direction')

OFFSETS = dict(N=(0, -1), E=(1, 0), W=(-1, 0), S=(0, 1))
MOVES = list('LRC')

def OOB(width, height, pt):
	return pt.x < 0 or width <= pt.x or pt.y < 0 or height <= pt.y

def DirPlusMove(dir, move):
	if move == 'C': return dir
	if move == 'L':
		return Point(x=dir.y, y=-dir.x)
	if move == 'R':
		return Point(x=-dir.y, y=dir.x)


def RandomPoint(width, height):
	return Point(x=random.randrange(width),
				 y=random.randrange(height))

def InitGame(width, height):
	head = Point(width / 2, height / 2)
	dir = Point(0, -1)
	return SnakeGame(width=width,
					 height=height,
					 state='LIVE',
					 fruit=RandomPoint(width, height),
					 head=head,
					 body=[Point(*(np.array(head) - dir))],
					 direction=dir)

def AdvanceGame(g, move):
	new_dir = DirPlusMove(g.direction, move)
	new_head = Point(*np.add(new_dir, g.head))
	
	if OOB(g.width, g.height, new_head):
		return g._replace(state='DEAD')
	
	if new_head == g.fruit:
		new_body = [g.head] + g.body
		new_fruit = RandomPoint(g.width, g.height)
		while new_fruit == new_head or new_fruit in new_body:
			new_fruit = RandomPoint(g.width, g.height)
	else:
		new_body = [g.head] + g.body[:-1]
		new_fruit = g.fruit

	if new_head in new_body:
		return g._replace(state='DEAD')
	return g._replace(head=new_head, body=new_body, fruit=new_fruit, direction=new_dir)

def ExportGame(stream, games):
	data = dict(width=games[0].width, height=games[0].height, steps=[])
	for g in games:	
		step = dict(pts=[], comment='')
		
		pts = [dict(g.fruit._asdict(), type='fruit')]
		pts += [dict(p._asdict(), type='snake')
				for p in [g.head] + g.body]
		data['steps'].append(dict(pts=pts, comment=str(get_stimuli(g))))
		
 	json.dump(data, stream, indent=2)
 	

def get_stimuli(g):
	l_eye = get_eye(g, DirPlusMove(g.direction, 'L'))
	r_eye = get_eye(g, DirPlusMove(g.direction, 'R'))
	c_eye = get_eye(g, DirPlusMove(g.direction, 'C'))
	return l_eye + c_eye + r_eye
	
def get_eye(g, dir):
	head = np.array(g.head)
	fruit_idx = -1
	wall_idx = -1
	body_idx = -1
	
	i = 1
	while True:
		p = Point(*(head + np.array(dir)*i))
		if OOB(g.width, g.height, p):
			wall_idx = i
			break
		if p == g.fruit:
			fruit_idx = i
		if p in g.body and body_idx == -1:
			body_idx = i
		i += 1
	
	return [wall_idx, fruit_idx, body_idx]

 	
if __name__ == '__main__':
	g = InitGame(11, 11)
	games = [g]

	tries = 0
	while len(games[-1].body) < 5:
		tries += 1
		g = InitGame(11, 11)
		games = [g]
		while g.state != 'DEAD':
			if len(games) > 100:
				break
			g = AdvanceGame(g, random.choice(MOVES))
			games.append(g)
	print 'tries = ', tries
	with open('ui/data.json', 'wb') as f:
		ExportGame(f, games)