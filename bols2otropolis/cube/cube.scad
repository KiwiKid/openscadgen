cube_size = [100,100,100];
cube_anchor = [0,0,0];
cube_spin = 0;
cube_orient = undef;
chamfer = 0;

include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

module cube_demo(){
	cuboid(size=cube_size, anchor=cube_anchor, spin=cube_spin, orient=cube_orient, chamfer=chamfer);
}

cube_demo();
