

	include <BOSL2/std.scad>;
include <BOSL2/walls.scad>;

	$fa = .01;

	$fs = $preview ? 5 : 1;
	$fn = 200;

    magnetHoles = "with-magnet-holes";
    magnetOffset = 6;
    plateBottom = 10;

	module phone_wedge(width=50){
		s = [[0, -78], [0, 40], [90, -25]];
        
        holderHeight = 80;
        holderWidth = 50;
        holderDepth = 5;
        holderFrameDepth = 25;
    
        magnetHoleDiameter = 6.0;
        magnetHoleDepth = 3.5;
        map_bottom_box_height = 3.5;
        rotate([0, 90, 0])
        zrot(60)
        difference(){
            union(){
                right(5)
                fwd(10)
                hex_panel(s, strut=1.5, spacing=15, h = width, frame = 0);
                 rotate([0,0,-35])
                 move([0,30,0])
                 union(){
                cuboid([holderHeight, holderDepth, holderWidth], rounding=2);
                   // back(1)
                 /*   difference(){
                    right(8)
                        cuboid([holderHeight*1.2, holderFrameDepth, holderWidth*1.2], rounding=1);
                        back(4)
                        #cuboid([80, holderFrameDepth, 50], rounding=2);
                    
                    }*/
                }
            }
            rotate([90,0,-35.0])
            move([-13,0,-20.9])
            union(){
            
             if(magnetHoles == "with-magnet-holes"){
                if(magnetOffset > 5){
                
                for (p = [[-magnetOffset, -magnetOffset], [magnetOffset, magnetOffset], [magnetOffset, -magnetOffset], [-magnetOffset, magnetOffset]]) {
                
                    translate([p.x, p.y, -plateBottom]) // Move to hole position
                        #cylinder(d=magnetHoleDiameter, h=magnetHoleDepth, anchor=CENTER); // Through-hole
                    
                }
                
                }else{
                    down(plateBottom)
                    cylinder(d=magnetHoleDiameter, h=magnetHoleDepth, anchor=CENTER); 
                }
	
    }
    }
    }
    }


    sliced(renderType="") {
        phone_wedge();
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

