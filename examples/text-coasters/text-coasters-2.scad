include <BOSL2/std.scad>

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 10;

coasterHeight = 10;
coasterRadius = 40;

textMainSize = 13;
textMainDepth = 8;
textContent = "LEFTY";
textSubContentSize = 2;
textSubContent = "Only Left Hand";
subContentStartAngle = 180;
subContentEndAngle = 360;

textAround = "LEFT LEFT LEFT LEFT LEFT LEFT LEFT LEFT LEFT LEFT LEFT LEFT L";

insetHeight = 6;
insetRadius = coasterRadius-4;

textAroundRadius = coasterRadius-2;

aroundTextStartAngle = 200;
aroundTextEndAngle = 370; 
aroundTextHeightOffset = 3;


stackingRimOffset = 1.5;
stackingRimHeight = 10;
stackingRimBaseRadius = -0.6;
stackingRimTopRadius = 3;


bottom_stacking_rim_z_offset = 2;

iconOffset = 16;
iconScale = 15;

difference() {
union(){
    cylinder(h=coasterHeight, r=coasterRadius);
    
    translate([0,0,coasterHeight+stackingRimOffset])
    #tube(h=stackingRimHeight, or1=coasterRadius+stackingRimBaseRadius, or2=coasterRadius+stackingRimTopRadius, wall=2, chamfer2=2, ichamfer1=2, ichamfer2=2, ochamfer1=2, ochamfer2=2);

}

    up(bottom_stacking_rim_z_offset)
    tube(h=stackingRimHeight, or1=coasterRadius+stackingRimBaseRadius, or2=coasterRadius+stackingRimTopRadius, wall=2, chamfer2=2, ichamfer1=2, ichamfer2=2, ochamfer1=2, ochamfer2=2);
    //translate([0,0,3])
    //#cylinder(h=coasterHeight, r=coasterRadius - insetHeight);

    
    // Text around the coaster
path = path3d(arc(100, r=textAroundRadius, angle=[0, 400]));
    color("red") stroke(path, width=.5);
    translate([0,0,aroundTextHeightOffset])
    !path_text(path, textAround , font="Liberation Mono",  size=5, center=true, lettersize = 4);
      /* 
    // Semi-Circle instructions
        path2 = path3d(arc(99, r=insetRadius, angle=[subContentStartAngle, subContentEndAngle]));
    color("red") stroke(path, width=.5);
    translate([0,0,insetHeight+2])
    path_text(path2, textSubContent, font="Liberation Mono", normal=UP, thickness=15, size=6, lettersize = 5/1.2, center=true);

   // Main Center Text
   
    translate([0,0,coasterHeight-textMainDepth])
    linear_extrude(height=textMainDepth+1)
    text(textContent, size=textMainSize, halign="center", valign="center");
    */
    
   /*
   translate([0,iconOffset,coasterHeight])
   scale(iconScale)
   difference(){
       import("./icon.stl");
       translate([0,4.1,0])
       #cuboid(5);
       
       translate([0,-3,0])
       #cuboid(5);
   }*/
     right(540)
    fwd(24)
    up(0)
    import("/Users/gregc/mine/making/3d-printing/openSCAD/openscadgen/examples/text-coasters/can.stl");
    
} 

// Text on the top center


  


    


