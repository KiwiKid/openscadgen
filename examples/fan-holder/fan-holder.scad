include <BOSL2/std.scad>

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;
renderType = "";

module fan_holder(coneHeight=80, bottomDiam=120, topDiam=60, centerRadius=37, centerDepth=40, textAroundOffset=35, aroundTextHeightOffset=-4, textAround="Sterling  Cumming", lockCicleRadius=4, lockCircleOffSet=-20, lockCicleDepth=2){

    difference() {
        cyl(h=coneHeight, d=bottomDiam, d2=topDiam, rounding1=2, rounding2=19);
        
        scale([1,1.1,1])
        up(centerDepth)
        cyl(h=100, d=centerRadius, rounding1=10);
        
        up(centerDepth-lockCicleDepth)
        left(centerRadius+lockCircleOffSet)
       cyl(h=100, d=lockCicleRadius);
          
        up(centerDepth-lockCicleDepth)     
       right(centerRadius+lockCircleOffSet)
       cyl(h=100, d=lockCicleRadius);




    // Text around the cone  
    path = path3d(arc(100, r=(topDiam-textAroundOffset), angle=[0, 360]));
    color("red") stroke(path, width=.5);
    translate([0,0,aroundTextHeightOffset])
    scale(2)
    down(4)
    path_text(path, textAround , font="Andale Mono:style=Regular",  size=10, center=true, lettersize =8, h=10, );
    }
}


sliceSize = 200;
offsetSize = 0;
horzOffset = 10;

  if(renderType == "horz-slice"){
      intersection(){
      fan_holder(); 
      fwd(sliceSize)
      left(sliceSize)
      up(horzOffset)
      #cuboid([1000,1000,0.3], anchor=[-1,-1,-1]);
      }
  }else if(renderType =="vert-slice"){
       intersection(){
          fan_holder();
          rotate([90,0,90])
          fwd(sliceSize)
          left(sliceSize)
          down(offsetSize)
          #cuboid([1000,1000,0.3], anchor=[-1,-1,-1]);
      }
  }else{ 
  
  fan_holder();
  }