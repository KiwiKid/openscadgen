

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

bottomRadius = 30;
topRadius = 10;
height = 100;

screwDiameter = 2.55;
screwHeadDiameter = 6;

YOffset=10;
XOffset=0;

screwOffset=4.2;

screwTranslate = [0,24,4];
holderTranslate = [XOffset,YOffset,height];
cutoutTranslate = [25,25,height/2];
cutoutRotate = [2,-5,0];
holderRadius = 4.5;

screwRotate = [0,0,0];

	module lamp_holder(bottomRadius=bottomRadius, topRadius=topRadius, height=height, screwDiameter=screwDiameter, screwHeadDiameter=screwHeadDiameter, XOffset=XOffset, YOffset=YOffset){
        difference(){
            hull(){
                cyl(r=bottomRadius, $fn=100);

                translate([XOffset, YOffset, height])
                #cyl(r=topRadius, $fn=100);
            }
            
            translate(holderTranslate)
            cyl(r=holderRadius, h=210);
            
            hull(){
            translate(holderTranslate)
            cyl(r=holderRadius*0.5, h=210);
            
            rotate(cutoutRotate)
            translate(cutoutTranslate)
            #cyl(r=1, h=150);

            }

            
            rotate(screwRotate)
            translate(screwTranslate)
            left(screwOffset)
            #cyl(d=screwDiameter, h=10);
                        
            rotate(screwRotate)
            translate(screwTranslate)
            right(screwOffset)
            #cyl(d=screwDiameter, h=10);


            
        }
	}


    sliced(renderType=renderType) {
        lamp_holder(bottomRadius=bottomRadius, topRadius=topRadius, height=height, screwDiameter=screwDiameter, screwHeadDiameter=screwHeadDiameter, XOffset=XOffset, YOffset=YOffset);
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 4,
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

