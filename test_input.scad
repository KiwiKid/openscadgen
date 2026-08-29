include <BOSL2/std.scad>;

module cubeDefault(size) { cuboid([size, size, size], anchor=[-1,-1,-1]); } cubeDefault(size=10);
