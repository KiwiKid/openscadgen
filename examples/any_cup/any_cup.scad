

	include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;

	/*
	renderType:
	use to print a test slice and confirm sizing before printing:
	 - "horzSlice" - horizontal slices (default)
	 - "vertSlice" - vertical slices
	 - "all" - the whole object
	*/
	renderType = "obj";

    wall_thickness = 3;
    radius = 52;
    height = 80;
    rounding = 4;
    chamfer = -10;
    chamfang = 80;
    ring_radius=radius;

    ring_radius_diff=7;
    
    // Text parameters
    text_string = "Van Royce  Van Royce  Van Royce";
    text_size = 10;
    text_thickness = 12.5;  // Thickness of the text
    text_offset = 0;  // Distance to shift letters outward from surface
    text_start_angle = 0;  // Starting angle for text path (degrees)
    text_angle_span = 360;  // Angle span for text path (degrees)
    text_z_pos = height * 0.3;  // Vertical position on cylinder
    kern = 1.7;

	module any_cup(radius=radius, height=height, wall_thickness=wall_thickness, rounding=rounding,chamfer=chamfer, chamfang=chamfang, ring_radius=ring_radius, ring_radius_diff=ring_radius_diff, text_string=text_string, text_size=text_size, text_thickness=text_thickness, text_offset=text_offset, text_start_angle=text_start_angle, text_angle_span=text_angle_span, text_z_pos=text_z_pos,kern=kern){
   
       difference(){
            union() {
		        cyl(h=height, r=radius+wall_thickness, rounding2=rounding, chamfer=chamfer, chamfang=chamfang);
                
                
            }
		down(wall_thickness)
		cyl(h=height, r=radius, rounding2=rounding, chamfer=chamfer, chamfang=chamfang);

        
        // Add text on outer surface using path_text
                up(text_z_pos) {
                    // Create 3D circular path at the outer radius
                    text_radius = radius + wall_thickness + text_offset;
                    path = path3d(arc(100, r=text_radius, angle=[text_start_angle, text_start_angle + text_angle_span]));
                    
                    // Create text along the path (path_text with 3D path creates 3D text)
                    path_text(
                        path, 
                        text_string, 
                        font="Arial:style=Bold", 
                        size=text_size, 
                        lettersize=text_size/1.2,
                        thickness=text_thickness,
                        offset=text_offset,
                        center=true,
                        kern=kern
                    );
                }
        }
        up(28)
        difference(){
                    cyl(h=height/4, r=ring_radius-1
                    +wall_thickness);
                    cyl(h=height/4, r=ring_radius
                    +wall_thickness-ring_radius_diff, chamfer=-8);
                }
	}

    sliced(renderType=renderType) {
        any_cup(radius=radius, height=height, wall_thickness=wall_thickness, rounding=rounding);
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 2,
    showRawSlices = false,
    horzSlicePos = [-500, -500, 0],
    vertSlicePos = [0, -500, -500]
) {
   
    module horz_slice(raw=false) {
        if (raw) {
            translate(horzSlicePos)
                cube([sliceSize, sliceSize, sliceThickness], center=false);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cube([sliceSize, sliceSize, sliceThickness], center=false);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cube([sliceThickness, sliceSize, sliceSize], center=false);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cube([sliceThickness, sliceSize, sliceSize], center=false);
            }
        }
    }

    if (renderType == "horzSlice") {
        horz_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "vertSlice") {
        vert_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "all") {
        // show raw slices for reference
        horz_slice(raw=true);
        vert_slice(raw=true);
        // show full object
        children();
    } else {
        // show full object
        children();
    }
}

