include <BOSL2/std.scad>

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

coasterHeight = 13;
coasterRadius = 33;

can_xy_scale = 1.001;

textTranslateZ = 9;
textMainSize = 9;
textMainDepth = 40;
textContent = "PHONE";
mainTextStrike = "with-strike";


textSubContentSize = 4;
textSubContent = "Only Left Hand HMMM";
textSubDepth = 6;
subContentStartAngle = 180;
subContentEndAngle = 360;

textSubYOffset = 10;

textAround = "LEFT LEFT LEFT LEFT LEFT LEFT LEFT LEFT";

insetHeight = 6;
insetRadius = coasterRadius-4;


aroundTextStartAngle = 200;
aroundTextEndAngle = 370; 
aroundTextHeightOffset = 10;


stackingRimOffset = 1.5;
stackingRimHeight = 30;
stackingRimBaseRadius = 10;
stackingRimTopRadius = 7;


textAroundRadiusDifference = 19.7; 
textAroundRadius = stackingRimBaseRadius+textAroundRadiusDifference;


bottom_stacking_rim_z_offset = -6;

iconOffset = 16;
iconScale = 15;
bottom_flatten_z_offset =-1;
bottom_flatten_size = [100,100,10];
chamfer=4;

module stacking_tube(){
 tube(h=stackingRimHeight, or1=coasterRadius+stackingRimBaseRadius, or2=coasterRadius+stackingRimBaseRadius, wall=15, rounding=10)
    fillet(h=2, r=2);
}

difference() {
union(){
    cylinder(h=coasterHeight, r=coasterRadius);
    
    translate([0,0,coasterHeight+stackingRimOffset])
   stacking_tube();
    

};

    up(bottom_stacking_rim_z_offset)
     stacking_tube();
     
    // Text around the coaster
   path = path3d(arc(100, r=textAroundRadius, angle=[0, 380]));
    color("red") stroke(path, width=.5);
    translate([0,0,aroundTextHeightOffset])
    scale(1.5)
    path_text(path, textAround , font="Liberation Mono",  size=6, center=true, lettersize =5, h=3);
 
   
    translate([0,0,textTranslateZ])
    linear_extrude(height=textMainDepth)
    text(textContent, size=textMainSize, halign="center", valign="center");
     
    if (mainTextStrike == "with-strike"){
        up(textTranslateZ+1.5)
        cuboid([coasterRadius*1.4,1.5,3]);
    }
    
    fwd(textSubYOffset)
    translate([0,0,textTranslateZ-1])
    linear_extrude(height=50+1)
    text(textSubContent, size=textSubContentSize, halign="center", valign="center");
   

     right(540)
    fwd(24)
    up(4)
    scale([can_xy_scale, can_xy_scale, 1])
    import("can.stl");
    
    
    // Bottom Flattener
    down(bottom_flatten_z_offset)
    #cuboid(bottom_flatten_size);
    
};


  

    















