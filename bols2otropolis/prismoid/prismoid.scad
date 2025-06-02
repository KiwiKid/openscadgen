prismoid_size1 = [35,50];
prismoid_size2 = [20,30];
prismoid_h = 20;
prismoid_shift = [0,0];
prismoid_xang = 0;
prismoid_yang = 0;
prismoid_rounding = 0;
prismoid_rounding1 = 0;
prismoid_rounding2 = 0;
prismoid_chamfer = 0;
prismoid_chamfer1 = 0;
prismoid_chamfer2 = 0;
prismoid_anchor = "BOTTOM";
prismoid_spin = 0;
prismoid_orient = "UP";

include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

module prismoid(){
	prismoid(size1=prismoid_size1, size2=prismoid_size2, h=prismoid_h, shift=prismoid_shift, xang=prismoid_xang, yang=prismoid_yang, rounding=prismoid_rounding, rounding1=prismoid_rounding1, rounding2=prismoid_rounding2, chamfer=prismoid_chamfer, chamfer1=prismoid_chamfer1, chamfer2=prismoid_chamfer2, anchor=prismoid_anchor, spin=prismoid_spin, orient=prismoid_orient);
}
