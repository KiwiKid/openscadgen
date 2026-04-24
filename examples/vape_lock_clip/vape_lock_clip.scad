
include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;
globalScale = 40;

clipWidth = 0.05;
clipLength = 0.3;
clipDepth = 000.02;
clipBarDepth = 0.04;
barSize = [clipWidth,0.02,clipBarDepth];

clipOffset1 = clipLength*0.5*0.7;
clipOffset2 = clipLength*0.5*0.93;

module vape_lock_clip(){
	cuboid([clipWidth, clipLength,clipDepth], rounding=0.02, edges="Z")
   
attach(TOP){
//up(clipHeight/2)
    fwd(clipOffset1)
    cuboid(barSize, rounding=0.005,  edges="Y");
    
   fwd(clipOffset2)
    cuboid(barSize, rounding=0.005, edges="Y");

    }
}

scale(globalScale)
vape_lock_clip();
