include <BOSL2/std.scad>;

module shape_with_curve() {
    straight_pts = [[0,0], [10,0], [10,15]];  // Define straight-line points
    
    // Bottom curve: Arc from (10,15) curving down to (0,10)
    bottom_curve = arc(n=10, r=10, angles=[0,90]);  
    bottom_curve = [for (p = bottom_curve) [p.x, p.y + 10]];

    // Top curve: Small rounded cap from (10,15) to (5,20) to (0,15)
    top_curve = arc(n=10, points=[[10,15], [5,20], [0,15]]);  

    // Create the final shape outline
    all_pts = concat(straight_pts, top_curve, reverse(bottom_curve)); 

    polygon(all_pts); // Create the 2D shape
}

// Toggle between thin or full rotation extrusion
thin_extrude = true;  

if (thin_extrude) {
    linear_extrude(height=2) shape_with_curve();  // Thin plate
} else {
    rotate_extrude() shape_with_curve();  // Full 3D object
}
//