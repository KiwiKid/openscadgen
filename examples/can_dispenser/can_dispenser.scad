
include <BOSL2/std.scad>;
include <BOSL2/cubetruss.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

rampScale = [1,1,0.4];
rampUpOffset = 1;

cubeSize = 25;
cubeDepthCount = 3;
mainLength = cubeDepthCount*cubeSize;
rampRoadSize = [cubeSize,mainLength,2];
canHoleBorder = 2;
canHoleWidth = 20;

upperRampHole = [cubeSize-canHoleBorder*2,canHoleWidth,10];
upperRampMove = [0,mainLength/2-upperRampHole[1]/2+-canHoleBorder, cubeSize];
module ramp(){
    fwd(mainLength/2)
    scale(rampScale)
    cubetruss_support(extents=cubeDepthCount, size=cubeSize, orient=BACK, strut=1.5, anchor=BOTTOM)
        attach(FWD){
    
    cuboid(rampRoadSize);
    }
    }

module can_dispenser(){

   difference(){
       union(){
        cubetruss(extents=[1,cubeDepthCount,2], bracing=true, size=cubeSize, strut=1.5, anchor=BOTTOM);
        
        
    // up(cubeSize+rampUpOffset)
       ramp();
       
       rotate([0,0,180]){
     up(cubeSize+rampUpOffset){
       ramp();
       }
       }
   
    }
   move(upperRampMove)
    #cuboid(upperRampHole);
   }
}



can_dispenser();
