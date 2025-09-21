

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

plugRadius = 14;
rimRadius = 26;
rimGapRadius = 3;


rimHeight = 24;
plugHeight = 70;
rimGapHeight = 30;

rounding=3;


path = ellipse(r=[20,10]);
tex = let(n=16,m=0.25) [
     [
         each resample_path(path3d(square(1)),n),
         each move([0.5,0.5],
             p=path3d(circle(d=0.5,$fn=n),m)),
         [1/2,1/2,0],
     ], [
         for (i=[0:1:n-1]) each [
             [i,(i+1)%n,(i+3)%n+n],
             [i,(i+3)%n+n,(i+2)%n+n],
             [2*n,n+i,n+(i+1)%n],
         ]
     ]
];

	module round_lid(plugRadius=plugRadius, plugHeight=plugHeight, rimRadius=rimRadius, rimHeight=rimHeight, rimGapHeight=rimGapHeight, rimGapRadius=rimGapRadius, rounding=rounding){
		cyl(h=plugHeight, r=plugRadius, rounding1=rounding);
        up(plugHeight/2-rimHeight/2+0.01)
        difference(){
            cyl(h=rimHeight, r=rimRadius+rimGapRadius, rounding=rounding, texture=tex);
            down(rimHeight/2)
            rotate([0,0,90])
            cyl(h=rimGapHeight, r=rimRadius, rounding=rounding);
        }
	}


    sliced(renderType=renderType) {
        round_lid( plugRadius=plugRadius, plugHeight=plugHeight, rimRadius=rimRadius, rimHeight=rimHeight, rimGapHeight=rimGapHeight, rimGapRadius=rimGapRadius);
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.1,
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

