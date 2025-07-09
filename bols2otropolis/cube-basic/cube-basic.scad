cube_size = [100,100,100];
cube_center = true;
cube_anchor = "CENTER";
cube_spin = 0;
cube_orient = undef;

include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

module cube_demo(){
	cube(size=cube_size, center=cube_center, anchor=cube_anchor, spin=cube_spin, orient=cube_orient);
}

cube_demo(); 