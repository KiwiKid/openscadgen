

	include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;


	module big_label(){
    
    name = "Diddys";
 text_angle = 10;
 base_width = !is_undef(base_width) ? base_width : 130 ;
base_size = [base_width, 30, 20];

text_size = 80;
text_height = 10;
include_base = "false";


            difference(){
            union(){



          //  rotate([-text_angle,0,0])
          //  cylindrical_extrude(or=140, ir=110)
          linear_extrude(height = text_height)
            text(text=name, size=text_size, halign="center", valign="center");
            
            
            if(include_base =="true"){
                fwd(60)
                up(93)
                cuboid(base_size, rounding=2);
                }
            }
            
            }
	}


    sliced(renderType="") {
        big_label();
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

