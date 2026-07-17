
include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;


pipeRadius = 5.5;

height = 80;

middleSectionGap = 20;
connectorOverlap = 25;
middleWallThickness = 2.5;

wallSize =0.8;
cut_direction = FRONT; // Try FRONT, LEFT, RIGHT, UP, or DOWN
cut_offset = 4;       // Distance to shift the cut plane from center
middleGapRadius = pipeRadius+middleWallThickness;

gripAmount = 2.2 ;

module middleGap(middleGapRadius=middleGapRadius,middleSectionGap=middleSectionGap){
//middle  blocker
    ycopies(3, sp=[0,0,0]){
        cyl(r=middleGapRadius, h=middleSectionGap);
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
   tube(middleSectionGap+connectorOverlap, ir=pipeRadius+wallSize, wall=middleWallThickness);
   
   
    middleGap(middleSectionGap=middleSectionGap);
    }
    
    

    
}

    half_of(cut_direction, cp=[0, cut_offset, 0]) {

pipe_straight_strengther();
}
