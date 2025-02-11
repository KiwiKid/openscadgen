$fn = 200;
 include <BOSL2/std.scad>


part = "map";

mapRotate = [0, -70, 0];


frameHeight = 82;
frameWidth = 82;
frameBorder =8;

if (part == "map" || part == "both"){
    translate([-55,-140,-113])
    rotate(mapRotate)
    difference(){
        import("./otago_peninsula_full.stl");
        translate([140, 140, 0])
    #rect_tube(size=[frameWidth+2,frameHeight+2], wall=frameBorder, rounding=10, ichamfer=7, h=10);
    }
}

 if(part == "stand" || part == "both") {
    // Create a cone using the cylinder function
    // A cone is a cylinder with a top radius of 0
    translate([0, 0, 4.8])
    cylinder(h = 55, r1 = 20, r2 = 0, center = true);   
   
   translate([-5,-frameHeight/2+41,20])
    rotate(mapRotate) 
    cube([frameWidth-frameBorder,frameHeight-frameBorder, 5], true);
    
    translate([-5, 0, 19])
    rotate(mapRotate)
    rect_tube(size=[frameWidth+2,frameHeight+2], wall=frameBorder, rounding=10, ichamfer=7, h=10);
}
// You can adjust the height (h), base radius (r1), and position as needed
