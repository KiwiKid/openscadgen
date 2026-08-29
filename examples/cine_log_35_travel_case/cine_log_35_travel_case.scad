

	include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;

    renderType = "vertSlice"; 
    caseScale = 1;


	module cine_log_35_travel_case(caseScale){
    scale(caseScale)
        union(){
		import("geprc-cinelog-30-v3-case-model_files/cinelog30v3boxlid.stl", anchor=CENTER);
        
        left(400)
        import("geprc-cinelog-30-v3-case-model_files/cinelog30v3boxclasp.stl", anchor=CENTER);
        
        right(400)
        import("geprc-cinelog-30-v3-case-model_files/cinelog30v3box.stl", anchor=CENTER);
        }

	}


    sliced(renderType) {
        cine_log_35_travel_case(caseScale);
    }
       








	
     
module sliced(
    renderType = "all",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1500,
    sliceThickness = 0.3,
    showRawSlices = false,
    horzSlicePos = [-600, -500, 0],
    horzSliceRotate = [0,0,90],
    vertSlicePos = [85, -600, -500],
    vertSliceRotate = [0,0,90]
) {
   
    module horz_slice(raw=false, horzSlicePos, horzSliceRotate) {
        if (raw) {
            rotate(horzSliceRotate)
            translate(horzSlicePos)
                cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                rotate(horzSliceRotate)
                translate(horzSlicePos)
                    cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
            }
        }
    }

    module vert_slice(raw=false,vertSlicePos, vertSliceRotate) {
        if (raw) {
            rotate(vertSliceRotate)
            translate(vertSlicePos)
                cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                   rotate(vertSliceRotate)
                translate(vertSlicePos)
                    cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
            }
        }
    }

    if (renderType == "horzSlice") {
        
        horz_slice(raw=showRawSlices, horzSlicePos, horzSliceRotate){
            children();
        }
    } else if (renderType == "vertSlice") {
    
     
        vert_slice(raw=showRawSlices, vertSlicePos, vertSliceRotate){
            children();
        }
    } else if (renderType == "all") {
        // show raw slices for reference
        horz_slice(raw=true, horzSlicePos, horzSliceRotate);
        vert_slice(raw=true, vertSlicePos, vertSliceRotate);
        // show full object
        children();
    } else {
        // show full object
        children();
    }
}

