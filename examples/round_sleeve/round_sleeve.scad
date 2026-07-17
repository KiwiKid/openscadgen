
include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;


sleeveRadius = 26.5;
sleeveHeight = 150;

sleeveWallWidth = .8;

gapWidth = 20;

module round_sleeve(){
difference(){
	cyl(r=sleeveRadius, h=sleeveHeight);
    
	cyl(r=sleeveRadius-sleeveWallWidth, h=sleeveHeight+1);
    
    fwd(sleeveRadius-sleeveWallWidth/2)
    cuboid([gapWidth,sleeveWallWidth*2,sleeveHeight+1]);
    }
}


round_sleeve();
