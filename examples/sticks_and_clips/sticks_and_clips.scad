
include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;


partType = "holder";// | "clip";

holderRadius = 5;
holeRadius = 1.68;

holderHeight =10;

stickXRotate = 30;
stickRotateOne = [30,stickXRotate,0];

heaps = 1000;

clipHeight = 3;
clipRadius = 10;
clipWall = 2;
clipScale = 1;

module clip(){
difference(){
    cyl(r=clipRadius, h=clipHeight, anchor=BOTTOM);
    
    cyl(r=clipRadius-clipWall, h=heaps);
    left(clipRadius)
    cuboid([9,110,100]);
    }
    }

module sticks_holder(){

    difference(){
        cyl(r=holderRadius,h=holderHeight, anchor=BOTTOM);

        cyl(r=holeRadius,h=heaps);
        
        left(holderRadius/2)
        rotate(stickRotateOne)
        cyl(r=holeRadius,h=heaps);
    }
    
    // Clip


}

if(partType == "holder" || partType == "all"){
    left(stickXRotate)
    sticks_holder();
}
if(partType == "clip" || partType == "all"){
     left(10*clipScale)
     scale(clipScale)
     clip();
 }

 
 
