

	include <BOSL2/std.scad>;
include <BOSL2/threading.scad>;

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
    
    base_radius_1 = 120; // 9.5
    base_radius_2 = 9; //90
    screw_radius = 8;
    base_height = 25;
    screw_height = 33;
    thread_pitch = 2.5;
    rounding_amount = 70;
    
    wall_width = 4;


	module globe_screw_on_base(base_radius_1=base_radius_1, base_radius_2=base_radius_2, screw_radius=screw_radius, base_height=base_height, screw_height=screw_height, thread_pitch=thread_pitch, rounding_amount=rounding_amount){
		difference() {
            union(){
                cyl(h=base_height, r=screw_radius+wall_width);
                
                tube(h=base_height, or1=base_radius_1, or2=screw_radius+5, ir1=base_radius_1-(wall_width*4), ir2=screw_radius);
            }
            // Threaded hole using BOSL2
            #threaded_rod(d=screw_radius*2, l=screw_height, pitch=thread_pitch, internal=true, $fn=32);
        }
	}


    sliced(renderType=renderType) {
        globe_screw_on_base();
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.3,
    showRawSlices = false,
    horzSlicePos = [-500, -500, 0],
    vertSlicePos = [0, -500, -500]
) {
   
    module horz_slice(raw=false) {
        if (raw) {
            translate(horzSlicePos)
                cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
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

