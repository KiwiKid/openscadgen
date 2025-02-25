include <BOSL2/std.scad>

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

coasterHeight = 6;
coasterRadius = 45;

textContentSize = 16;
textContent = "LEFTY";
textSubContentSize = 3;
textSubContent = "Only Drink with Left Hand";
subContentStartAngle = 180;
subContentEndAngle = 360;

textAround = "LEFT LEFT LEFT LEFT LEFT LEF LEFT LEFT LEFT LEFT LEFT LEFT LEFT LEFT";

insetHeight = 6;
insetRadius = coasterRadius-3;

textAroundRadius = coasterRadius;

aroundTextStartAngle = 200;
aroundTextEndAngle = 370; 
aroundTextHeightOffset = 0.5;

stackingRimOffset = 1.5;
stackingRimHeight = 7;
stackingRimBaseRadius = -0.6;
stackingRimTopRadius = 3;

difference() {
union(){
    cylinder(h=coasterHeight, r=coasterRadius);
    
    translate([0,0,coasterHeight+stackingRimOffset])
    tube(h=stackingRimHeight, or1=coasterRadius+stackingRimBaseRadius, or2=coasterRadius+stackingRimTopRadius, wall=2, chamfer2=2, ichamfer1=2, ichamfer2=2, ochamfer1=2, ochamfer2=2);

}
    //translate([0,0,3])
    //#cylinder(h=coasterHeight, r=coasterRadius - insetHeight);

    
    // Text around the coaster
   path = path3d(arc(100, r=textAroundRadius, angle=[0, 360]));
    color("red") stroke(path, width=.5);
    translate([0,0,aroundTextHeightOffset])
    path_text(path,textAround , font="Liberation Mono", thickness=2, size=5, center=true, lettersize = 4);
    
    // Semi-Circle instructions
        path2 = path3d(arc(99, r=insetRadius, angle=[subContentStartAngle, subContentEndAngle]));
    color("red") stroke(path, width=.5);
    translate([0,0,insetHeight+2])
    path_text(path2, textSubContent, font="Liberation Mono", normal=UP, thickness=15, size=6, lettersize = 5/1.2, center=true);

   
          translate([0,0,1])
    linear_extrude(height=10)
   // wrapped_text(textContent, max_width=50, size=10);
    text(textContent, size=textContentSize, halign="center", valign="center");
    
   
    
}


// Text on the top center


