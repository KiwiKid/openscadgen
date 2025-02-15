$fn = 200;
include <BOSL2/std.scad>;

part = "both";  // "map", "stand", or "both"

mapRotate = [0, -70, 0];
mapSTLScale = 0.34; //0.96
mapSTLScaling = [mapSTLScale, mapSTLScale, mapSTLScale];

frameWidth = 82;
frameHeight = 82;
frameBorder = 8;
frameDepth = 10;
frameRounding=10;

place = "dunedin";

map_stl_translate = [0,0,0];
map_stl_file = "./banks_peninsula.stl";

if (place == "dunedin") {  
// otago_peninsula_full.stl banks-peninsula.stl banks-peninsula
map_stl_file = "otago_peninsula_full.stl";  // Ensure the STL is in the same directory
map_stl_x_translate = -204;
map_stl_y_translate = 5;
map_stl_z_translate = -5;
map_stl_translate = [map_stl_x_translate, map_stl_y_translate, map_stl_z_translate];
} else if(place == "chch") {
// otago_peninsula_full.stl banks-peninsula.stl banks-peninsula
map_stl_file = "./banks_peninsula.stl";  // Ensure the STL is in the same directory
map_stl_x_translate = -204;
map_stl_y_translate = 5;
map_stl_z_translate = -5;
map_stl_translate = [map_stl_x_translate, map_stl_y_translate, map_stl_z_translate];
mapRotate = [0, -50, 0];


} else {
    assert(false, "Select value place");
}



translate([0, 0, 0])  // Adjust to center the map properly
difference() {
    // Import and align STL inside frame
    translate(map_stl_translate)  // Slightly inset to sit flush
    rotate(mapRotate)
    scale(mapSTLScaling)
    #import(map_stl_file, center=true);

    // Create the frame around the STL
    rect_tube(size=[frameWidth, frameHeight], wall=frameBorder, rounding=frameRounding, ichamfer=7, h=frameDepth, anchor=CENTER);
}

