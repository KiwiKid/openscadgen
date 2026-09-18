
// file::///./config.toml
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
innerRadius = 4;

cylHeight = 20;
wallSize = -1;// 1;

outerRadius = innerRadius + wallSize;

cupInnerRoundiung = 2.5;
cupOuterRounding1 = 0;
cupOuterChamfer1 = -6.5;
starStepCount = 3;

floorDepth = 1;
subMode = "sizer";


mode =  "nailHook-slice";// "decagon";// "dowelPattern"; "nailHook";


 nailHookRadius = 10;
innerFn = 5;
nailHookHeight = 150;
nailHookBottomHoleHeight = 100;
nailHookBottomHoleRadius = 4;
    
nailHookTopHoleRadius = 3;

module decagon(){
    difference(){
            cyl(r=outerRadius, h=cylHeight);
            
                up(wallSize)
                linear_extrude(h=cylHeight)
            regular_ngon(n=6, ir=innerRadius);
          }
}


module cylBoid(r=5, h=100, cylChamfer1=6, cylChamfer2=2, anchor=BOT, innerFn=5){
    intersection(){
        cuboid([r, r, h], chamfer=3, anchor=BOT);
        union(){
        cyl(r=r, h=h, chamfer1=cylChamfer1, chamfer2=cylChamfer2, anchor=anchor);
        
        
        cyl(r=r*3, h=20, anchor=anchor);
        }
       }

}

module nailHook(){
        
    //left(10)    cylBoid();
        
    //nailhook
   rotate([90,90,0])
        difference(){
         cylBoid(r=nailHookRadius, h=nailHookHeight, cylChamfer1=-5, cylChamfer2=1, anchor=BOT);
           // cyl(r=nailHookHoleRadius, h=nailHookHeight, chamfer1=cupOuterChamfer1, chamfer=-3, chamfer2=3, anchor=BOT);
            down(0.001)
                    cyl(r=nailHookBottomHoleRadius, h=nailHookBottomHoleHeight, chamfer1=-2, anchor=BOT, $fn=innerFn);
                    up(nailHookBottomHoleHeight-1)
                    cyl(r=nailHookTopHoleRadius, h=nailHookBottomHoleHeight+1, anchor=BOT, $fn=innerFn);
            }

}




	module simple_cylinder(mode = "tube"){
    if(mode == "tube"){
        difference(){
            cyl(r=outerRadius, h=cylHeight);
            
           cyl(r=innerRadius, h=cylHeight+0.1);
           }
    } else if(mode == "decagon"){
    
    if(subMode == "sizer"){
    intersection(){
        decagon();
        up(5)
        cuboid([100,100,1]);
        }
     } else {
        decagon();
     }
       
    }else if(mode =="cup") {
            difference(){
                cyl(r=outerRadius, h=cylHeight);
                
                up(wallSize)
                cyl(r=innerRadius, h=cylHeight+0.1, rounding=cupInnerRoundiung);
           }    
        
	} else  if(mode =="dowelPattern") {
            difference(){
               
                cyl(r=outerRadius, h=cylHeight, chamfer1=cupOuterChamfer1, chamfer2=-1);
                
 //               / Inner cutout with vertical dowel-like ribs lifted by 10 units
            up(wallSize+floorDepth) {
                linear_sweep(
                    star(r=innerRadius, n=24, step=starStepCount), 
                    h=cylHeight+0.1, 
                    center=true
                );
                down(wallSize+1)
                cyl(r=innerRadius*0.9, h=cylHeight-wallSize-floorDepth+0.001, rounding=cupInnerRoundiung, chamfer2=-1);
                }
           }    
        
	} else if(mode =="dowel") {

                linear_sweep(
                    star(r=innerRadius, n=24, step=starStepCount), 
                    h=cylHeight+0.1, 
                    center=true
                );
                down(wallSize+1)
                cyl(r=innerRadius*0.96, h=cylHeight*1.1, rounding=cupInnerRoundiung);
                
           
        
	} else if(mode == "nailHook") {
    nailHook();
                
    } else if(mode == "nailHook-slice") {
        intersection(){
            //up(nailHookBottomHoleHeight-0.1)
            fwd(nailHookBottomHoleHeight)
            down(20)
            cuboid([100,10,100], anchor=BOT);
           nailHook();
        }
                
    } else {

            assert(false, "should have a valid mode ");
        }
        }
    
    
    simple_cylinder(mode=mode);

