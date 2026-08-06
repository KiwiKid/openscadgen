
include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;


pipeRadius = 6.7;

height = 100;

middleSectionGap = 60;
connectorOverlap = 35;
middleWallThickness = 1;

wallSize =1.5;
cut_direction = FRONT; // Try FRONT, LEFT, RIGHT, UP, or DOWN
cut_offset = 5;       // Distance to shift the cut plane from center
middleGapRadius = 9;

gripAmount = 2.2 ;

module middleGap(middleGapRadius=middleGapRadius,middleSectionGap=middleSectionGap){
//middle  blocker
    ycopies(1, sp=[0,0,0], n=10){
        cyl(r=middleGapRadius, h=middleSectionGap, chamfer=5);
    }
    }

module pipe_straight_strengther(){
difference(){
	tube(height, or=pipeRadius+wallSize, ir=pipeRadius);
    
    middleGap(middleSectionGap=middleSectionGap);
    
    //left blocker
  //  back(pipeRadius*gripAmount)
  //  cuboid([20,20,100]);
    }
    
    

    difference(){
   tube(middleSectionGap+connectorOverlap, ir=pipeRadius+wallSize, or=middleGapRadius+wallSize, ochamfer=7);
   
   
    middleGap(middleSectionGap=middleSectionGap);
    }
    
    

    
}

    half_of(cut_direction, cp=[0, cut_offset, 0]) {

pipe_straight_strengther();
}
