

	include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;


	module window_holder_cover(){

        bottom_height = 10;
        
        bottom_primoid_size = [70,60];
        holder_offset = 10;
        
        cube_size = [35,50,20];
        cube_wall= 5;
        

        difference(){
            prismoid(size1=bottom_primoid_size, size2=[cube_size[0],cube_size[1]], h=20);
            
            down(cube_wall)
            prismoid(size1=bottom_primoid_size, size2=[cube_size[0],cube_size[1]+10], h=20);
        }

        up(bottom_height*2)
        difference(){
            cuboid(cube_size, anchor=BOTTOM);
            
            up(holder_offset)
            #cuboid(cube_size-[cube_wall, -15, 0], anchor=BOTTOM);
        }
        
	}


    sliced(renderType="") {
        window_holder_cover();
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

