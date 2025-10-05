include <BOSL2/std.scad>;
$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;


topHeight = 25;
topOffset = 1.95;

bottomOffset=9;
bottomHeight=12;
clyRadius = 1.4;

cuboid([30,10,2], rounding=5, edges="Z")

down(bottomHeight/2)
left(bottomOffset)
cyl(h=bottomHeight, r=clyRadius, rounding1=clyRadius);

down(bottomHeight/2)
right(bottomOffset)
cyl(h=bottomHeight, r=clyRadius, rounding1=clyRadius);

up(topHeight/2)
left(topOffset)
cyl(h=topHeight, r=clyRadius, rounding2=clyRadius);

up(topHeight/2)
right(topOffset)
cyl(h=topHeight, r=clyRadius, rounding2=clyRadius);
