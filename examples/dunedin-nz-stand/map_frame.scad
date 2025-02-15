$fn = 200;
include <BOSL2/std.scad>;

part = "both";  // "map", "stand", or "both"

mapRotate = [0, -70, 0];
mapSTLScale = 0.37; //0.96
mapSTLScaling = [mapSTLScale, mapSTLScale, mapSTLScale];

frameWidth = 82;
frameHeight = 82;
frameBorder = 8;
frameDepth = 3;

map_stl_file = "banks-peninsula.stl";  // Ensure the STL is in the same directory


if (part == "map" || part == "both") {
    translate([0, 0, 0])  // Adjust to center the map properly
    difference() {
        // Import and align STL inside frame
        translate([-226, 5, -5])  // Slightly inset to sit flush
      //  rotate(mapRotate)
        scale(mapSTLScaling)
        import(map_stl_file, center=true);

        // Create the frame around the STL
        rect_tube(size=[frameWidth, frameHeight], wall=frameBorder, rounding=10, ichamfer=7, h=frameDepth, anchor=CENTER);
    }
}
