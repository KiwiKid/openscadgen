include <BOSL2/std.scad>;
$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

baseLength = 125;
baseHeight = 140;

cutoutHeight = 110;
cutoutDepth = 55;
cutoutBottomOffset = -3;

arcAngleSharpness = 0.25;

module shape_with_curve() {
    straight_pts = [
        [baseLength, 0], 
        [0, baseHeight], 
    ];  

    // Create a smooth top arc from (0, baseHeight) to (cutoutDepth, cutoutHeight)
    top_arc = arc(n=20, points=[[0, baseHeight], [1,baseHeight-arcAngleSharpness],  [baseLength, 0]]);

    lower_pts = [
      //  [cutoutDepth,cutoutHeight],
        [0, 0],
        [cutoutDepth+cutoutBottomOffset, 0],
        [cutoutDepth, cutoutHeight],
        [0, cutoutHeight],
        [0, baseHeight],
    ];  

    polygon(concat(straight_pts, top_arc, lower_pts));  // Combine all points
}

thin_extrude = true;  // Set to false for full rotate_extrude

if (thin_extrude) {
    linear_extrude(height=1) shape_with_curve();  // Thin extrusion (plate-like)
} else {
    rotate_extrude() shape_with_curve();  // Full revolve (3D shape)
}
