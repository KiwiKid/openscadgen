include <BOSL2/std.scad>;

$fn = 100;

motorBarRadius= 6/2;
motorBarNotchRadius = 1.6;

railWheelHeight = 20;


outerRadiusRounding= 7;
outerRadius= 18;

innerCutoutRadius = 10;
innerCutoutHeight = 14;

maxOuterRadius = 45/2;

motorBodyDepth = 24;
motorBodyWidth = 51;
motorBodyWidthClearance = 2;
motorBodyDepthClearance = 2;
mode = "strap"; //all, wheel, strap
strapWall = 3;
strapWidth = motorBodyWidth+motorBodyWidthClearance;


strapSize = [strapWidth,motorBodyDepth+railWheelHeight+motorBodyDepthClearance];
strapHeight = 10;

strapGapWidth = strapWidth-30;
setScrewRadius = 1;
module notched_cyl(){
     intersection(){
         cyl(r=motorBarRadius, h=railWheelHeight+0.001, chamfer2=-1);
         fwd(motorBarNotchRadius/2)
         cuboid([100,motorBarRadius*2, 100]);
     }
}


module curtain_string_rail_big_wheel_v2(){
    difference(){
        intersection(){
        cyl(r=outerRadius, h=railWheelHeight, rounding=-outerRadiusRounding);
         cyl(r=maxOuterRadius, h=railWheelHeight);
        }
        
        tube(ir=innerCutoutRadius, h=innerCutoutHeight, or=maxOuterRadius)
        notched_cyl();
      
        down(railWheelHeight/2-notchHeight/2+0.001)
        xrot(90)
     strap_notch();
        
        ycyl(r=setScrewRadius,h=40, anchor=TOP+FWD);
        
    }
        
}

notchHeight = 2;

module strap_notch(notchRadiusScale=1){
     ycyl(r1=5*notchRadiusScale, r2=3*notchRadiusScale, h=notchHeight);
}

module curtain_string_rail_big_wheel_strap(){

     difference(){
         rect_tube(
          size = strapSize,
          wall = strapWall,
          h = strapHeight,
          rounding = 3
        );
        cuboid([strapGapWidth,100,100], anchor=FWD);
         
     }
        fwd(railWheelHeight/2)
    
     
     fwd(strapSize[1]/2-railWheelHeight/2-strapWall-notchHeight/2)
     up(strapHeight/2)
     strap_notch(notchRadiusScale=0.8);
    
}


if(mode == "strap" || mode == "all"){
    xrot(-90)
    curtain_string_rail_big_wheel_strap();
}
if(mode == "wheel" || mode == "all"){

curtain_string_rail_big_wheel_v2();
}